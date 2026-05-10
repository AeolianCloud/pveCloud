<script setup lang="ts">
import { NCard, NTabPane, NTabs } from 'naive-ui'
import { computed, onMounted, reactive, ref } from 'vue'

import {
  createProduct,
  createProductPlan,
  createNetworkType,
  createSalesRegion,
  createServerOsTemplate,
  deleteNetworkType,
  deleteProduct,
  deleteProductPlan,
  deleteSalesRegion,
  deleteServerOsTemplate,
  getProductPlans,
  getProducts,
  getNetworkTypes,
  getPlanNetworkTypes,
  getPlanOsTemplates,
  getPlanPrices,
  getPlanRegions,
  getSalesRegions,
  getServerOsTemplates,
  updateNetworkType,
  updatePlanNetworkTypes,
  updatePlanOsTemplates,
  updatePlanPrices,
  updatePlanRegions,
  updateProduct,
  updateProductPlan,
  updateProductPlanStatus,
  updateProductStatus,
  updateSalesRegion,
  updateServerOsTemplate,
  type NetworkTypeItem,
  type NetworkTypePayload,
  type PlanPriceItem,
  type PlanPricePayload,
  type ProductItem,
  type ProductPayload,
  type ProductPlanItem,
  type ProductPlanPayload,
  type SalesRegionItem,
  type SalesRegionPayload,
  type ServerOsTemplateItem,
  type ServerOsTemplatePayload,
} from '../../api/product-catalog'
import { confirm, message } from '../../utils/feedback'
import PlanPricesDialog from './components/PlanPricesDialog.vue'
import PlanRelationsDialog from './components/PlanRelationsDialog.vue'
import ProductEditorDialog from './components/ProductEditorDialog.vue'
import ProductPlanEditorDialog from './components/ProductPlanEditorDialog.vue'
import ProductPlansTab from './components/ProductPlansTab.vue'
import ProductsTab from './components/ProductsTab.vue'
import NetworkTypeEditorDialog from './components/NetworkTypeEditorDialog.vue'
import NetworkTypesTab from './components/NetworkTypesTab.vue'
import SalesRegionEditorDialog from './components/SalesRegionEditorDialog.vue'
import SalesRegionsTab from './components/SalesRegionsTab.vue'
import ServerOsTemplateEditorDialog from './components/ServerOsTemplateEditorDialog.vue'
import ServerOsTemplatesTab from './components/ServerOsTemplatesTab.vue'
import type { DialogMode, ProductCatalogTabKey } from './types'

const productList = ref<ProductItem[]>([])
const planList = ref<ProductPlanItem[]>([])
const regionList = ref<SalesRegionItem[]>([])
const templateList = ref<ServerOsTemplateItem[]>([])
const networkTypeList = ref<NetworkTypeItem[]>([])
const planPriceMap = reactive<Record<number, PlanPriceItem[]>>({})
const planRegionMap = reactive<Record<number, SalesRegionItem[]>>({})
const planTemplateMap = reactive<Record<number, ServerOsTemplateItem[]>>({})
const planNetworkTypeMap = reactive<Record<number, NetworkTypeItem[]>>({})

const activeTabs = ref<ProductCatalogTabKey>('products')
const loading = reactive({ products: false, plans: false, regions: false, templates: false, networkTypes: false })

const productDialogVisible = ref(false)
const planDialogVisible = ref(false)
const regionDialogVisible = ref(false)
const templateDialogVisible = ref(false)
const networkTypeDialogVisible = ref(false)
const relationDialogVisible = ref(false)
const priceDialogVisible = ref(false)

const productDialogMode = ref<DialogMode>('create')
const planDialogMode = ref<DialogMode>('create')
const regionDialogMode = ref<DialogMode>('create')
const templateDialogMode = ref<DialogMode>('create')
const networkTypeDialogMode = ref<DialogMode>('create')

const productForm = reactive<ProductPayload>({ type: 'server', slug: '', name: '', summary: '', description: '', status: 'draft', visible: true, sort_order: 0 })
const productFormId = ref<number | null>(null)

const planForm = reactive<ProductPlanPayload>({ product_id: 0, code: '', name: '', summary: '', cpu_cores: 2, memory_mb: 2048, system_disk_gb: 50, data_disk_gb: 0, bandwidth_mbps: 100, traffic_gb: null, public_ip_count: 1, virtualization: 'kvm', architecture: 'x86_64', is_featured: false, status: 'draft', visible: true, sort_order: 0 })
const planFormId = ref<number | null>(null)

const regionForm = reactive<SalesRegionPayload>({ code: '', name: '', country: '', city: '', summary: '', status: 'active', visible: true, sort_order: 0 })
const regionFormId = ref<number | null>(null)

const templateForm = reactive<ServerOsTemplatePayload>({ code: '', name: '', os_family: 'linux', distribution: '', version: '', architecture: 'x86_64', summary: '', status: 'active', visible: true, sort_order: 0 })
const templateFormId = ref<number | null>(null)

const networkTypeForm = reactive<NetworkTypePayload>({ code: '', name: '', summary: '', status: 'active', visible: true, sort_order: 0 })
const networkTypeFormId = ref<number | null>(null)

const relationTargetPlan = ref<ProductPlanItem | null>(null)
const selectedRegionIds = ref<number[]>([])
const selectedTemplateIds = ref<number[]>([])
const selectedNetworkTypeIds = ref<number[]>([])

const priceTargetPlan = ref<ProductPlanItem | null>(null)
const priceForm = reactive<PlanPricePayload[]>(makePriceForm())

const productsById = computed(() => Object.fromEntries(productList.value.map((item) => [item.id, item])))

function makePriceForm(): PlanPricePayload[] {
  return [
    { billing_cycle: 'monthly', price_cents: 0, original_price_cents: null, currency: 'CNY', status: 'active', sort_order: 10 },
    { billing_cycle: 'quarterly', price_cents: 0, original_price_cents: null, currency: 'CNY', status: 'inactive', sort_order: 20 },
    { billing_cycle: 'semi_yearly', price_cents: 0, original_price_cents: null, currency: 'CNY', status: 'inactive', sort_order: 30 },
    { billing_cycle: 'yearly', price_cents: 0, original_price_cents: null, currency: 'CNY', status: 'inactive', sort_order: 40 },
  ]
}

function resetProductForm() {
  productFormId.value = null
  Object.assign(productForm, { type: 'server', slug: '', name: '', summary: '', description: '', status: 'draft', visible: true, sort_order: 0 })
}

function resetPlanForm() {
  planFormId.value = null
  Object.assign(planForm, { product_id: productList.value[0]?.id || 0, code: '', name: '', summary: '', cpu_cores: 2, memory_mb: 2048, system_disk_gb: 50, data_disk_gb: 0, bandwidth_mbps: 100, traffic_gb: null, public_ip_count: 1, virtualization: 'kvm', architecture: 'x86_64', is_featured: false, status: 'draft', visible: true, sort_order: 0 })
}

function resetRegionForm() {
  regionFormId.value = null
  Object.assign(regionForm, { code: '', name: '', country: '', city: '', summary: '', status: 'active', visible: true, sort_order: 0 })
}

function resetTemplateForm() {
  templateFormId.value = null
  Object.assign(templateForm, { code: '', name: '', os_family: 'linux', distribution: '', version: '', architecture: 'x86_64', summary: '', status: 'active', visible: true, sort_order: 0 })
}

function resetNetworkTypeForm() {
  networkTypeFormId.value = null
  Object.assign(networkTypeForm, { code: '', name: '', summary: '', status: 'active', visible: true, sort_order: 0 })
}

async function loadAll() {
  loading.products = true
  loading.plans = true
  loading.regions = true
  loading.templates = true
  loading.networkTypes = true
  try {
    const [products, plans, regions, templates, networkTypes] = await Promise.all([
      getProducts({ per_page: 100 }),
      getProductPlans({ per_page: 100 }),
      getSalesRegions(),
      getServerOsTemplates(),
      getNetworkTypes(),
    ])
    productList.value = products.list
    planList.value = plans.list
    regionList.value = regions
    templateList.value = templates
    networkTypeList.value = networkTypes
    await Promise.all(planList.value.map((plan) => loadPlanRelations(plan.id)))
  } catch (error) {
    message.error(error instanceof Error ? error.message : '产品目录加载失败')
  } finally {
    loading.products = false
    loading.plans = false
    loading.regions = false
    loading.templates = false
    loading.networkTypes = false
  }
}

async function loadPlanRelations(planId: number) {
  const [prices, regions, templates, networkTypes] = await Promise.all([
    getPlanPrices(planId),
    getPlanRegions(planId),
    getPlanOsTemplates(planId),
    getPlanNetworkTypes(planId),
  ])
  planPriceMap[planId] = prices
  planRegionMap[planId] = regions
  planTemplateMap[planId] = templates
  planNetworkTypeMap[planId] = networkTypes
}

function openCreateProduct() {
  resetProductForm()
  productDialogMode.value = 'create'
  productDialogVisible.value = true
}

function openEditProduct(item: ProductItem) {
  productDialogMode.value = 'edit'
  productFormId.value = item.id
  Object.assign(productForm, { type: 'server', slug: item.slug, name: item.name, summary: item.summary || '', description: item.description || '', status: item.status, visible: item.visible, sort_order: item.sort_order })
  productDialogVisible.value = true
}

async function saveProduct() {
  if (productDialogMode.value === 'create') {
    await createProduct(productForm)
  } else if (productFormId.value != null) {
    await updateProduct(productFormId.value, productForm)
  }
  message.success('已保存产品')
  productDialogVisible.value = false
  await loadAll()
}

async function toggleProductStatus(item: ProductItem) {
  const nextStatus = item.status === 'active' ? 'inactive' : 'active'
  await updateProductStatus(item.id, nextStatus)
  message.success('状态已更新')
  await loadAll()
}

async function removeProduct(item: ProductItem) {
  if (!(await confirmDelete(`确认删除产品"${item.name}"吗？存在套餐时后端会拒绝删除。`))) return
  await deleteProduct(item.id)
  message.success('产品已删除')
  await loadAll()
}

function openCreatePlan() {
  resetPlanForm()
  planDialogMode.value = 'create'
  planDialogVisible.value = true
}

function openEditPlan(item: ProductPlanItem) {
  planDialogMode.value = 'edit'
  planFormId.value = item.id
  Object.assign(planForm, { product_id: item.product_id, code: item.code, name: item.name, summary: item.summary || '', cpu_cores: item.cpu_cores, memory_mb: item.memory_mb, system_disk_gb: item.system_disk_gb, data_disk_gb: item.data_disk_gb, bandwidth_mbps: item.bandwidth_mbps, traffic_gb: item.traffic_gb, public_ip_count: item.public_ip_count, virtualization: 'kvm', architecture: 'x86_64', is_featured: item.is_featured, status: item.status as ProductPlanPayload['status'], visible: item.visible, sort_order: item.sort_order })
  planDialogVisible.value = true
}

async function savePlan() {
  if (planDialogMode.value === 'create') {
    await createProductPlan(planForm)
  } else if (planFormId.value != null) {
    await updateProductPlan(planFormId.value, planForm)
  }
  message.success('已保存套餐')
  planDialogVisible.value = false
  await loadAll()
}

function openCreateRegion() {
  resetRegionForm()
  regionDialogMode.value = 'create'
  regionDialogVisible.value = true
}

function openEditRegion(item: SalesRegionItem) {
  regionDialogMode.value = 'edit'
  regionFormId.value = item.id
  Object.assign(regionForm, { code: item.code, name: item.name, country: item.country || '', city: item.city || '', summary: item.summary || '', status: item.status as SalesRegionPayload['status'], visible: item.visible, sort_order: item.sort_order })
  regionDialogVisible.value = true
}

async function saveRegion() {
  if (regionDialogMode.value === 'create') {
    await createSalesRegion(regionForm)
  } else if (regionFormId.value != null) {
    await updateSalesRegion(regionFormId.value, regionForm)
  }
  message.success('已保存地域')
  regionDialogVisible.value = false
  await loadAll()
}

function openCreateTemplate() {
  resetTemplateForm()
  templateDialogMode.value = 'create'
  templateDialogVisible.value = true
}

function openEditTemplate(item: ServerOsTemplateItem) {
  templateDialogMode.value = 'edit'
  templateFormId.value = item.id
  Object.assign(templateForm, { code: item.code, name: item.name, os_family: item.os_family as ServerOsTemplatePayload['os_family'], distribution: item.distribution, version: item.version, architecture: 'x86_64', summary: item.summary || '', status: item.status as ServerOsTemplatePayload['status'], visible: item.visible, sort_order: item.sort_order })
  templateDialogVisible.value = true
}

async function saveTemplate() {
  if (templateDialogMode.value === 'create') {
    await createServerOsTemplate(templateForm)
  } else if (templateFormId.value != null) {
    await updateServerOsTemplate(templateFormId.value, templateForm)
  }
  message.success('已保存模板')
  templateDialogVisible.value = false
  await loadAll()
}

function openCreateNetworkType() {
  resetNetworkTypeForm()
  networkTypeDialogMode.value = 'create'
  networkTypeDialogVisible.value = true
}

function openEditNetworkType(item: NetworkTypeItem) {
  networkTypeDialogMode.value = 'edit'
  networkTypeFormId.value = item.id
  Object.assign(networkTypeForm, { code: item.code, name: item.name, summary: item.summary || '', status: item.status as NetworkTypePayload['status'], visible: item.visible, sort_order: item.sort_order })
  networkTypeDialogVisible.value = true
}

async function saveNetworkType() {
  if (networkTypeDialogMode.value === 'create') {
    await createNetworkType(networkTypeForm)
  } else if (networkTypeFormId.value != null) {
    await updateNetworkType(networkTypeFormId.value, networkTypeForm)
  }
  message.success('已保存网络类型')
  networkTypeDialogVisible.value = false
  await loadAll()
}

async function openPriceDialog(item: ProductPlanItem) {
  priceTargetPlan.value = item
  const list = planPriceMap[item.id] || []
  const fresh = makePriceForm()
  fresh.forEach((entry) => {
    const found = list.find((price) => price.billing_cycle === entry.billing_cycle)
    if (found) {
      entry.price_cents = found.price_cents
      entry.original_price_cents = found.original_price_cents
      entry.currency = found.currency as 'CNY'
      entry.status = found.status as 'active' | 'inactive'
      entry.sort_order = found.sort_order
    }
  })
  priceForm.splice(0, priceForm.length, ...fresh)
  priceDialogVisible.value = true
}

async function savePrices() {
  if (!priceTargetPlan.value) return
  await updatePlanPrices(priceTargetPlan.value.id, priceForm)
  message.success('已保存价格')
  priceDialogVisible.value = false
  await loadPlanRelations(priceTargetPlan.value.id)
}

async function openRelationDialog(item: ProductPlanItem) {
  relationTargetPlan.value = item
  selectedRegionIds.value = (planRegionMap[item.id] || []).map((region) => region.id)
  selectedTemplateIds.value = (planTemplateMap[item.id] || []).map((template) => template.id)
  selectedNetworkTypeIds.value = (planNetworkTypeMap[item.id] || []).map((networkType) => networkType.id)
  relationDialogVisible.value = true
}

async function saveRelations() {
  if (!relationTargetPlan.value) return
  await Promise.all([
    updatePlanRegions(relationTargetPlan.value.id, selectedRegionIds.value),
    updatePlanOsTemplates(relationTargetPlan.value.id, selectedTemplateIds.value),
    updatePlanNetworkTypes(relationTargetPlan.value.id, selectedNetworkTypeIds.value),
  ])
  message.success('已保存关联')
  relationDialogVisible.value = false
  await loadPlanRelations(relationTargetPlan.value.id)
}

async function togglePlanStatus(item: ProductPlanItem) {
  const nextStatus = item.status === 'active' ? 'inactive' : 'active'
  await updateProductPlanStatus(item.id, nextStatus)
  message.success('状态已更新')
  await loadAll()
}

async function removePlan(item: ProductPlanItem) {
  if (!(await confirmDelete(`确认删除套餐"${item.name}"吗？套餐价格和关联配置会一并删除。`))) return
  await deleteProductPlan(item.id)
  message.success('套餐已删除')
  await loadAll()
}

async function removeRegion(item: SalesRegionItem) {
  if (!(await confirmDelete(`确认删除销售地域"${item.name}"吗？仍被套餐关联时后端会拒绝删除。`))) return
  await deleteSalesRegion(item.id)
  message.success('销售地域已删除')
  await loadAll()
}

async function removeTemplate(item: ServerOsTemplateItem) {
  if (!(await confirmDelete(`确认删除系统模板"${item.name}"吗？仍被套餐关联时后端会拒绝删除。`))) return
  await deleteServerOsTemplate(item.id)
  message.success('系统模板已删除')
  await loadAll()
}

async function removeNetworkType(item: NetworkTypeItem) {
  if (!(await confirmDelete(`确认删除网络类型"${item.name}"吗？仍被套餐关联时后端会拒绝删除。`))) return
  await deleteNetworkType(item.id)
  message.success('网络类型已删除')
  await loadAll()
}

async function confirmDelete(content: string) {
  try {
    await confirm({ title: '确认删除', content, type: 'warning', positiveText: '删除' })
    return true
  } catch {
    return false
  }
}

function planPublishIssues(item: ProductPlanItem) {
  const issues: string[] = []
  const product = productsById.value[item.product_id]
  if (!product || product.status !== 'active' || !product.visible) issues.push('产品未公开')
  if (!['active', 'sold_out'].includes(item.status) || !item.visible) issues.push('套餐未公开')
  if (!(planPriceMap[item.id] || []).some((price) => price.status === 'active')) issues.push('缺启用价格')
  if ((planRegionMap[item.id] || []).length === 0) issues.push('缺销售地域')
  if ((planTemplateMap[item.id] || []).length === 0) issues.push('缺系统模板')
  if ((planNetworkTypeMap[item.id] || []).length === 0) issues.push('缺网络类型')
  return issues
}

function statusLabel(status: string) {
  const labels: Record<string, string> = {
    active: '上架',
    draft: '草稿',
    inactive: '下架',
    sold_out: '售罄',
  }
  return labels[status] || status
}

function statusTagType(status: string) {
  if (status === 'active') return 'success'
  if (status === 'sold_out') return 'warning'
  if (status === 'inactive') return 'info'
  return 'default'
}

onMounted(loadAll)
</script>

<template>
  <section class="products-page">
    <div class="section-pad">
      <NCard>
        <NTabs v-model:value="activeTabs" type="line" animated>
          <NTabPane name="products" tab="产品">
            <ProductsTab
              :products="productList"
              :loading="loading.products"
              :status-label="statusLabel"
              :status-tag-type="statusTagType"
              @create="openCreateProduct"
              @edit="openEditProduct"
              @toggle-status="toggleProductStatus"
              @delete="removeProduct"
            />
          </NTabPane>

          <NTabPane name="plans" tab="套餐">
            <ProductPlansTab
              :plans="planList"
              :products-by-id="productsById"
              :loading="loading.plans"
              :plan-publish-issues="planPublishIssues"
              :status-label="statusLabel"
              :status-tag-type="statusTagType"
              @create="openCreatePlan"
              @edit="openEditPlan"
              @prices="openPriceDialog"
              @relations="openRelationDialog"
              @toggle-status="togglePlanStatus"
              @delete="removePlan"
            />
          </NTabPane>

          <NTabPane name="regions" tab="地域">
            <SalesRegionsTab
              :regions="regionList"
              :loading="loading.regions"
              :status-label="statusLabel"
              :status-tag-type="statusTagType"
              @create="openCreateRegion"
              @edit="openEditRegion"
              @delete="removeRegion"
            />
          </NTabPane>

          <NTabPane name="templates" tab="系统模板">
            <ServerOsTemplatesTab
              :templates="templateList"
              :loading="loading.templates"
              :status-label="statusLabel"
              :status-tag-type="statusTagType"
              @create="openCreateTemplate"
              @edit="openEditTemplate"
              @delete="removeTemplate"
            />
          </NTabPane>

          <NTabPane name="networkTypes" tab="网络类型">
            <NetworkTypesTab
              :network-types="networkTypeList"
              :loading="loading.networkTypes"
              :status-label="statusLabel"
              :status-tag-type="statusTagType"
              @create="openCreateNetworkType"
              @edit="openEditNetworkType"
              @delete="removeNetworkType"
            />
          </NTabPane>
        </NTabs>
      </NCard>
    </div>

    <ProductEditorDialog
      :visible="productDialogVisible"
      :mode="productDialogMode"
      :form="productForm"
      @update:visible="productDialogVisible = $event"
      @save="saveProduct"
    />
    <ProductPlanEditorDialog
      :visible="planDialogVisible"
      :mode="planDialogMode"
      :form="planForm"
      :products="productList"
      @update:visible="planDialogVisible = $event"
      @save="savePlan"
    />
    <SalesRegionEditorDialog
      :visible="regionDialogVisible"
      :mode="regionDialogMode"
      :form="regionForm"
      @update:visible="regionDialogVisible = $event"
      @save="saveRegion"
    />
    <ServerOsTemplateEditorDialog
      :visible="templateDialogVisible"
      :mode="templateDialogMode"
      :form="templateForm"
      @update:visible="templateDialogVisible = $event"
      @save="saveTemplate"
    />
    <NetworkTypeEditorDialog
      :visible="networkTypeDialogVisible"
      :mode="networkTypeDialogMode"
      :form="networkTypeForm"
      @update:visible="networkTypeDialogVisible = $event"
      @save="saveNetworkType"
    />
    <PlanPricesDialog
      :visible="priceDialogVisible"
      :target-plan="priceTargetPlan"
      :prices="priceForm"
      @update:visible="priceDialogVisible = $event"
      @save="savePrices"
    />
    <PlanRelationsDialog
      :visible="relationDialogVisible"
      v-model:selected-region-ids="selectedRegionIds"
      v-model:selected-template-ids="selectedTemplateIds"
      v-model:selected-network-type-ids="selectedNetworkTypeIds"
      :target-plan="relationTargetPlan"
      :regions="regionList"
      :templates="templateList"
      :network-types="networkTypeList"
      @update:visible="relationDialogVisible = $event"
      @save="saveRelations"
    />
  </section>
</template>

<style scoped>
.products-page {
  padding-bottom: 32px;
}
</style>
