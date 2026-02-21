// internal/middleware/oplog.go
// 操作日志中间件，挂在需要审计的写操作路由上，自动记录执行人、模块、动作和结果。
package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// WriteOpLog 返回一个写操作日志的中间件。
//
// module：操作模块（如 "admin"、"role"）
// action：操作动作（如 "create"、"update"、"delete"、"set_status"）
//
// 记录策略：
// - 同时记录“成功”和“失败”（失败包括：HTTP 4xx/5xx 或业务 code != 0）
// - target_id：优先从 URL :id 解析；创建类接口会尝试从响应 data.id 解析
// - target_label：
//   - 优先从 URL :id 对应的记录里预查询（例如 admin 删除前先查 username）
//   - 创建类接口会尝试从请求体里提取（例如 admin.create 取 username、role.create 取 label/name）
//
// 注意：本中间件会读取并恢复 Request Body（用于提取 target_label），不会影响 handler 的 ShouldBindJSON。
func WriteOpLog(db *gorm.DB, module, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ---- 1) 请求信息预处理：提取 target_id / target_label（尽量在执行前拿到） ----

		// target_id：从 URL 参数 :id 取，取不到则为 0
		var targetID uint
		if idStr := c.Param("id"); idStr != "" && idStr != "/" {
			var n uint64
			if _, err := parseUint(idStr, &n); err == nil {
				targetID = uint(n)
			}
		}

		// 读取并恢复请求 body（最多读取 1MB，避免异常大 body 影响性能/内存）
		var reqBody []byte
		if c.Request != nil && c.Request.Body != nil {
			limited := io.LimitReader(c.Request.Body, 1<<20)
			b, _ := io.ReadAll(limited)
			_ = c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewReader(b))
			reqBody = b
		}

		// target_label：优先通过 DB 预查询（适用于 update/delete 等）
		targetLabel := prefetchTargetLabel(db, module, targetID)
		// 创建类接口（通常没有 :id），从请求体尝试提取目标描述
		if targetLabel == "" && targetID == 0 {
			targetLabel = extractTargetLabelFromBody(module, action, reqBody)
		}

		// ---- 2) 捕获响应：用于判断业务 code 是否成功，以及创建类接口提取 data.id ----
		w := &captureWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil)}
		c.Writer = w

		// ---- 3) 执行后续 handler ----
		c.Next()

		// ---- 4) 解析响应，判断本次操作是否成功 ----
		httpStatus := c.Writer.Status()
		bizCode := 0

		var env responseEnvelope
		if json.Unmarshal(w.body.Bytes(), &env) == nil {
			bizCode = env.Code
		}

		// status：1 成功  0 失败（失败包括 HTTP 4xx/5xx 或业务 code != 0）
		opStatus := int8(1)
		if httpStatus >= 400 || bizCode != 0 {
			opStatus = 0
		}

		// 创建类接口：若 target_id 仍为空，尝试从响应 data.id 提取
		if targetID == 0 && len(env.Data) > 0 {
			if id := extractIDFromData(env.Data); id > 0 {
				targetID = id
			}
		}

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(uint)
		uname, _ := username.(string)

		log := model.AdminOpLog{
			AdminUserID: uid,
			Username:    uname,
			Module:      module,
			Action:      action,
			TargetID:    targetID,
			TargetLabel: targetLabel,
			Status:      opStatus,
			IP:          c.ClientIP(),
		}
		// 静默写入，失败不影响业务
		db.Create(&log)
	}
}

// responseEnvelope 操作日志中间件使用的响应包裹结构。
// 与后端统一响应格式保持一致：{ code, message, data }。
type responseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// captureWriter 捕获响应 body，用于解析业务 code / data。
type captureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *captureWriter) Write(b []byte) (int, error) {
	// 先捕获，再写入真实响应
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// prefetchTargetLabel 尝试通过 DB 预查询 target_label（适合 update/delete）。
func prefetchTargetLabel(db *gorm.DB, module string, targetID uint) string {
	if targetID == 0 {
		return ""
	}
	switch module {
	case "admin":
		var username string
		// 只查询必要字段，避免把整行数据加载出来
		_ = db.Table("admin_users").Select("username").Where("id = ?", targetID).Scan(&username).Error
		return username
	case "role":
		var label string
		_ = db.Table("admin_roles").Select("label").Where("id = ?", targetID).Scan(&label).Error
		return label
	default:
		return ""
	}
}

// extractTargetLabelFromBody 从请求体 JSON 中提取目标描述。
// 说明：这里不追求覆盖所有情况，只覆盖最常用模块，保证日志里“看得懂”。
func extractTargetLabelFromBody(module, action string, body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return ""
	}

	// admin:create 通常包含 username
	if module == "admin" && action == "create" {
		if v, ok := m["username"].(string); ok {
			return v
		}
	}

	// role:create 通常包含 label/name
	if module == "role" && action == "create" {
		if v, ok := m["label"].(string); ok && v != "" {
			return v
		}
		if v, ok := m["name"].(string); ok {
			return v
		}
	}

	return ""
}

// extractIDFromData 尝试从响应 data 中提取 id 字段（用于创建类接口补 target_id）。
func extractIDFromData(data json.RawMessage) uint {
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return 0
	}
	// JSON number 会被解成 float64
	if v, ok := m["id"].(float64); ok && v > 0 {
		return uint(v)
	}
	return 0
}

// parseUint 简单解析无符号整数，避免引入 strconv 包外的依赖。
func parseUint(s string, out *uint64) (int, error) {
	var v uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s}
		}
		v = v*10 + uint64(c-'0')
	}
	*out = v
	return len(s), nil
}

type parseError struct{ s string }

func (e *parseError) Error() string { return "invalid number: " + e.s }
