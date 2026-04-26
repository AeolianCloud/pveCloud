<template>
  <Card class="admin-table-panel">
    <template #content>
      <slot />
    </template>
    <template v-if="$slots.footer" #footer>
      <footer class="admin-table-panel-footer">
        <slot name="footer" />
      </footer>
    </template>
  </Card>
</template>

<style scoped>
.admin-table-panel {
  overflow: hidden;
}

.admin-table-panel :deep(.p-card-body) {
  gap: 0;
  padding: 0;
}

.admin-table-panel-footer {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  color: var(--muted);
}

@media (max-width: 760px) {
  .admin-table-panel-footer {
    align-items: flex-start;
    flex-direction: column;
    padding-block: 12px;
  }
}
</style>
