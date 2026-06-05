<template>
  <div class="tab-bar">
    <button
      v-for="cat in categories"
      :key="cat.id"
      class="tab-btn"
      :class="{ active: activeCategory === cat.id }"
      @click="$emit('select', cat.id)"
    >
      {{ cat.icon }} {{ cat.label }}
    </button>
  </div>
</template>
<script setup>
defineProps({
  categories: { type: Array, required: true },
  activeCategory: { type: String, required: true }
})
defineEmits(['select'])
</script>
<style scoped>
.tab-bar { display: flex; border-bottom: 1px solid #27272a; flex-wrap: wrap; }
.tab-btn { flex: 1; padding: 0.75rem 0.5rem; background: transparent; border: none; border-bottom: 2px solid transparent; color: #71717a; font-size: 0.8125rem; font-family: 'Inter', sans-serif; font-weight: 500; cursor: pointer; transition: all 0.2s; white-space: nowrap; }
.tab-btn:hover { color: #d4d4d8; background: rgba(255,255,255,0.02); }
.tab-btn.active { color: #e4e4e7; border-bottom-color: #8b5cf6; }
</style>
