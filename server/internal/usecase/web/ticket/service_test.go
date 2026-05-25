package ticket

import (
	"context"
	"testing"

	"gorm.io/gorm"

	domainticket "github.com/AeolianCloud/pveCloud/server/internal/domain/ticket"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestDetailScopesToCurrentUserAndPublicTicketData(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicket(t, db, domainticket.StatusWaitingUser)

	service := NewService(db, storageConfigForTicketTest(t))
	detail, err := service.Detail(context.Background(), 11, "TIC-USER-1")
	if err != nil {
		t.Fatalf("load own ticket detail: %v", err)
	}
	if detail.TicketNo != "TIC-USER-1" || detail.Title != "Need help" {
		t.Fatalf("unexpected ticket detail: %#v", detail.TicketItem)
	}
	if len(detail.Tags) != 1 || detail.Tags[0].Name != "Public Tag" || detail.Tags[0].Visibility != "public" {
		t.Fatalf("web detail should include only public tags, got %#v", detail.Tags)
	}
	if len(detail.Messages) != 1 || detail.Messages[0].SenderType != domainticket.SenderUser || detail.Messages[0].Content != "first message" {
		t.Fatalf("web detail should return user-visible message timeline, got %#v", detail.Messages)
	}

	_, err = service.Detail(context.Background(), 12, "TIC-USER-1")
	if apperrors.From(err).Code != apperrors.ErrNotFound.Code {
		t.Fatalf("cross-user ticket detail should be hidden as not found, got %v", err)
	}
}

func TestReplyRejectsClosedTicketWithoutWritingMessage(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicket(t, db, domainticket.StatusClosed)

	service := NewService(db, storageConfigForTicketTest(t))
	_, err := service.Reply(context.Background(), 11, "TIC-USER-1", webdto.TicketMessageRequest{Content: "new reply"}, nil)
	if apperrors.From(err).Code != apperrors.ErrConflict.Code {
		t.Fatalf("reply to closed ticket should conflict, got %v", err)
	}

	var messageCount int64
	if err := db.Table("ticket_messages").Where("ticket_id = ?", 100).Count(&messageCount).Error; err != nil {
		t.Fatalf("count ticket messages: %v", err)
	}
	if messageCount != 1 {
		t.Fatalf("closed ticket reply must not write a new message, got %d", messageCount)
	}
}

func TestCreateLinksOwnOrderWithoutInstance(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	seedWebTicketOrder(t, db, 200, 11, "ORD-USER-ONLY")

	service := NewService(db, storageConfigForTicketTest(t))
	detail, err := service.Create(context.Background(), 11, webdto.TicketCreateRequest{
		Title:    "Order question",
		Category: "order",
		Priority: "normal",
		Content:  "Please help check this order.",
		OrderNo:  "ORD-USER-ONLY",
	}, nil)
	if err != nil {
		t.Fatalf("create order-linked ticket: %v", err)
	}
	if detail.OrderNo == nil || *detail.OrderNo != "ORD-USER-ONLY" || detail.InstanceNo != nil {
		t.Fatalf("ticket should keep order-only link, got order=%v instance=%v", detail.OrderNo, detail.InstanceNo)
	}

	var row struct {
		OrderID    *uint64
		InstanceID *uint64
	}
	if err := db.Table("tickets").Select("order_id, instance_id").Where("ticket_no = ?", detail.TicketNo).Take(&row).Error; err != nil {
		t.Fatalf("load order-linked ticket: %v", err)
	}
	if row.OrderID == nil || *row.OrderID != 200 || row.InstanceID != nil {
		t.Fatalf("ticket should persist only order id, got %#v", row)
	}
}

func TestCreateLinksOwnInstanceAndBackfillsOrder(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	seedWebTicketOrder(t, db, 201, 11, "ORD-USER-1")
	seedWebTicketInstance(t, db, 301, 11, 201, "ORD-USER-1", "INS-USER-1")

	service := NewService(db, storageConfigForTicketTest(t))
	detail, err := service.Create(context.Background(), 11, webdto.TicketCreateRequest{
		Title:      "Instance unreachable",
		Category:   "technical",
		Priority:   "normal",
		Content:    "Please help check my instance.",
		InstanceNo: "INS-USER-1",
	}, nil)
	if err != nil {
		t.Fatalf("create linked ticket: %v", err)
	}
	if detail.OrderNo == nil || *detail.OrderNo != "ORD-USER-1" || detail.InstanceNo == nil || *detail.InstanceNo != "INS-USER-1" {
		t.Fatalf("ticket should expose linked order and instance, got order=%v instance=%v", detail.OrderNo, detail.InstanceNo)
	}

	var row struct {
		OrderID    *uint64
		OrderNo    *string
		InstanceID *uint64
		InstanceNo *string
	}
	if err := db.Table("tickets").Select("order_id, order_no, instance_id, instance_no").Where("ticket_no = ?", detail.TicketNo).Take(&row).Error; err != nil {
		t.Fatalf("load created ticket link: %v", err)
	}
	if row.OrderID == nil || *row.OrderID != 201 || row.InstanceID == nil || *row.InstanceID != 301 {
		t.Fatalf("ticket should persist server-resolved link ids, got %#v", row)
	}
}

func TestCreateRejectsOtherUserInstance(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	seedWebTicketUser(t, db, 12, "other-user", "other@example.com")
	seedWebTicketOrder(t, db, 202, 12, "ORD-OTHER-1")
	seedWebTicketInstance(t, db, 302, 12, 202, "ORD-OTHER-1", "INS-OTHER-1")

	service := NewService(db, storageConfigForTicketTest(t))
	_, err := service.Create(context.Background(), 11, webdto.TicketCreateRequest{
		Title:      "Cross user instance",
		Category:   "technical",
		Priority:   "normal",
		Content:    "This should be rejected.",
		InstanceNo: "INS-OTHER-1",
	}, nil)
	if apperrors.From(err).Code != apperrors.ErrNotFound.Code {
		t.Fatalf("cross-user instance should be hidden as not found, got %v", err)
	}

	var ticketCount int64
	if err := db.Table("tickets").Count(&ticketCount).Error; err != nil {
		t.Fatalf("count tickets: %v", err)
	}
	if ticketCount != 0 {
		t.Fatalf("rejected ticket creation must not write tickets, got %d", ticketCount)
	}
}

func TestCreateRejectsMismatchedOrderAndInstance(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	seedWebTicketOrder(t, db, 203, 11, "ORD-USER-1")
	seedWebTicketOrder(t, db, 204, 11, "ORD-USER-2")
	seedWebTicketInstance(t, db, 303, 11, 204, "ORD-USER-2", "INS-USER-2")

	service := NewService(db, storageConfigForTicketTest(t))
	_, err := service.Create(context.Background(), 11, webdto.TicketCreateRequest{
		Title:      "Mismatched link",
		Category:   "technical",
		Priority:   "normal",
		Content:    "Order and instance do not match.",
		OrderNo:    "ORD-USER-1",
		InstanceNo: "INS-USER-2",
	}, nil)
	if apperrors.From(err).Code != apperrors.ErrValidation.Code {
		t.Fatalf("mismatched order and instance should be validation error, got %v", err)
	}
}

func TestListFiltersByInstanceNo(t *testing.T) {
	db := openWebTicketDB(t)
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	instanceNo := "INS-FILTER-1"
	if err := db.Exec(`
INSERT INTO tickets (
  id, ticket_no, user_id, category, priority, title, status, instance_no,
  last_message_at, last_user_message_at
) VALUES
  (401, 'TIC-MATCH', 11, 'technical', 'normal', 'Match', ?, ?, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3)),
  (402, 'TIC-OTHER', 11, 'technical', 'normal', 'Other', ?, 'INS-FILTER-2', CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))`,
		domainticket.StatusWaitingAdmin, instanceNo, domainticket.StatusWaitingAdmin,
	).Error; err != nil {
		t.Fatalf("insert filtered tickets: %v", err)
	}

	service := NewService(db, storageConfigForTicketTest(t))
	page, err := service.List(context.Background(), 11, webdto.TicketListQuery{InstanceNo: instanceNo})
	if err != nil {
		t.Fatalf("list tickets by instance: %v", err)
	}
	if page.Total != 1 || len(page.List) != 1 || page.List[0].TicketNo != "TIC-MATCH" || page.List[0].InstanceNo == nil || *page.List[0].InstanceNo != instanceNo {
		t.Fatalf("unexpected filtered result: %#v", page)
	}
}

func storageConfigForTicketTest(t *testing.T) config.StorageConfig {
	t.Helper()
	return config.StorageConfig{Driver: "local", LocalPath: t.TempDir(), MaxSize: 1024 * 1024, AllowedTypes: []string{"text/plain"}}
}

func openWebTicketDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		webTicketUsersSchema,
		webTicketAdminUsersSchema,
		webTicketOrdersSchema,
		webTicketInstancesSchema,
		webTicketTicketsSchema,
		webTicketMessagesSchema,
		webTicketFileAttachmentsSchema,
		webTicketMessageAttachmentsSchema,
		webTicketTagsSchema,
		webTicketTagBindingsSchema,
		webTicketUserBusinessLogsSchema,
	)
	return db
}

func seedWebTicket(t *testing.T, db *gorm.DB, status string) {
	t.Helper()
	seedWebTicketUser(t, db, 11, "ticket-user", "ticket@example.com")
	seedWebTicketUser(t, db, 12, "other-user", "other@example.com")
	if err := db.Exec(`INSERT INTO admin_users (id, username, email, display_name, password_hash, status) VALUES (?, ?, ?, ?, ?, ?)`, 21, "admin", "admin@example.com", "Admin", "hash", "active").Error; err != nil {
		t.Fatalf("insert admin user: %v", err)
	}
	if err := db.Exec(`
INSERT INTO tickets (
  id, ticket_no, user_id, category, priority, title, status, assignee_admin_id,
  last_message_at, last_user_message_at, first_response_due_at, resolution_due_at
) VALUES (100, 'TIC-USER-1', 11, 'technical', 'normal', 'Need help', ?, 21, CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3), CURRENT_TIMESTAMP(3))`, status).Error; err != nil {
		t.Fatalf("insert ticket: %v", err)
	}
	if err := db.Exec(`INSERT INTO ticket_messages (id, ticket_id, sender_type, sender_user_id, content) VALUES (200, 100, ?, 11, 'first message')`, domainticket.SenderUser).Error; err != nil {
		t.Fatalf("insert ticket message: %v", err)
	}
	if err := db.Exec(`INSERT INTO ticket_tags (id, name, visibility, status, sort_order) VALUES (300, 'Public Tag', 'public', 'active', 1), (301, 'Internal Tag', 'internal', 'active', 2)`).Error; err != nil {
		t.Fatalf("insert ticket tags: %v", err)
	}
	if err := db.Exec(`INSERT INTO ticket_tag_bindings (ticket_id, tag_id) VALUES (100, 300), (100, 301)`).Error; err != nil {
		t.Fatalf("insert ticket tag bindings: %v", err)
	}
}

func seedWebTicketUser(t *testing.T, db *gorm.DB, id uint64, username string, email string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, id, username, email, "hash", "active").Error; err != nil {
		t.Fatalf("insert user %d: %v", id, err)
	}
}

func seedWebTicketOrder(t *testing.T, db *gorm.DB, id uint64, userID uint64, orderNo string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO orders (id, order_no, user_id) VALUES (?, ?, ?)`, id, orderNo, userID).Error; err != nil {
		t.Fatalf("insert order %s: %v", orderNo, err)
	}
}

func seedWebTicketInstance(t *testing.T, db *gorm.DB, id uint64, userID uint64, orderID uint64, orderNo string, instanceNo string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO instances (id, instance_no, user_id, order_id, order_no) VALUES (?, ?, ?, ?, ?)`, id, instanceNo, userID, orderID, orderNo).Error; err != nil {
		t.Fatalf("insert instance %s: %v", instanceNo, err)
	}
}

const webTicketUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketAdminUsersSchema = `
CREATE TABLE admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NULL,
  display_name VARCHAR(64) NOT NULL DEFAULT '',
  password_hash VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_orders_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketInstancesSchema = `
CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  instance_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_instances_instance_no (instance_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketTicketsSchema = `
CREATE TABLE tickets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NULL,
  order_no VARCHAR(64) NULL,
  instance_id BIGINT UNSIGNED NULL,
  instance_no VARCHAR(64) NULL,
  category VARCHAR(32) NOT NULL,
  priority VARCHAR(32) NOT NULL,
  title VARCHAR(160) NOT NULL,
  status VARCHAR(32) NOT NULL,
  assignee_admin_id BIGINT UNSIGNED NULL,
  assigned_by_admin_id BIGINT UNSIGNED NULL,
  assigned_at DATETIME(3) NULL,
  last_message_at DATETIME(3) NOT NULL,
  last_user_message_at DATETIME(3) NULL,
  last_admin_message_at DATETIME(3) NULL,
  first_response_due_at DATETIME(3) NULL,
  first_responded_at DATETIME(3) NULL,
  resolution_due_at DATETIME(3) NULL,
  resolved_at DATETIME(3) NULL,
  closed_by_type VARCHAR(32) NULL,
  closed_by_user_id BIGINT UNSIGNED NULL,
  closed_by_admin_id BIGINT UNSIGNED NULL,
  closed_at DATETIME(3) NULL,
  close_reason VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_tickets_ticket_no (ticket_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketMessagesSchema = `
CREATE TABLE ticket_messages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket_id BIGINT UNSIGNED NOT NULL,
  sender_type VARCHAR(32) NOT NULL,
  sender_user_id BIGINT UNSIGNED NULL,
  sender_admin_id BIGINT UNSIGNED NULL,
  content TEXT NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketFileAttachmentsSchema = `
CREATE TABLE file_attachments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  original_name VARCHAR(255) NOT NULL,
  storage_path VARCHAR(500) NOT NULL,
  mime_type VARCHAR(128) NOT NULL,
  extension VARCHAR(32) NOT NULL,
  size BIGINT UNSIGNED NOT NULL,
  sha256 VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketMessageAttachmentsSchema = `
CREATE TABLE ticket_message_attachments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket_id BIGINT UNSIGNED NOT NULL,
  message_id BIGINT UNSIGNED NOT NULL,
  file_id BIGINT UNSIGNED NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketTagsSchema = `
CREATE TABLE ticket_tags (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(40) NOT NULL,
  color VARCHAR(32) NULL,
  visibility VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_by_admin_id BIGINT UNSIGNED NULL,
  updated_by_admin_id BIGINT UNSIGNED NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketTagBindingsSchema = `
CREATE TABLE ticket_tag_bindings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket_id BIGINT UNSIGNED NOT NULL,
  tag_id BIGINT UNSIGNED NOT NULL,
  created_by_admin_id BIGINT UNSIGNED NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webTicketUserBusinessLogsSchema = `
CREATE TABLE user_business_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  username VARCHAR(64) NULL,
  email VARCHAR(191) NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(96) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  summary VARCHAR(500) NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
