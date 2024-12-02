<template>
  <div class="list-view">
    <div class="list-header">
      <h1>{{ title }}</h1>
      <div class="list-info">
        <slot name="header-info"></slot>
        <div v-if="showSearch" class="search-box">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="searchPlaceholder"
            @input="handleSearch"
          />
        </div>
      </div>
    </div>

    <div class="list-content">
      <table class="list-table">
        <thead>
          <tr>
            <th
              v-for="column in columns"
              :key="column.key"
              :class="{ sortable: column.sortable }"
              @click="column.sortable ? handleSort(column.key) : null"
            >
              {{ column.title }}
              <span v-if="column.sortable && sortBy === column.key" class="sort-icon">
                {{ sortOrder === 'desc' ? '↓' : '↑' }}
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <template v-if="!loading">
            <slot name="table-rows" :columns="columns"></slot>
          </template>
          <tr v-else>
            <td :colspan="columns.length" class="loading">
              <div class="loading-spinner"></div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="pagination">
      <div class="pagination-info">
        共 {{ total }} 条记录
        <select v-model="currentPageSize" @change="handleSizeChange" class="page-size-select">
          <option v-for="size in pageSizes" :key="size" :value="size">{{ size }} 条/页</option>
        </select>
      </div>
      <div class="pagination-buttons">
        <button @click="handlePageChange(1)" :disabled="currentPage === 1" class="page-btn">
          首页
        </button>
        <button
          @click="handlePageChange(currentPage - 1)"
          :disabled="currentPage === 1"
          class="page-btn"
        >
          上一页
        </button>
        <div class="page-numbers">
          <button
            v-for="pageNum in displayedPages"
            :key="pageNum"
            @click="handlePageChange(pageNum)"
            :class="['page-btn', { active: currentPage === pageNum }]"
          >
            {{ pageNum }}
          </button>
        </div>
        <button
          @click="handlePageChange(currentPage + 1)"
          :disabled="currentPage === totalPages"
          class="page-btn"
        >
          下一页
        </button>
        <button
          @click="handlePageChange(totalPages)"
          :disabled="currentPage === totalPages"
          class="page-btn"
        >
          末页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { debounce } from 'lodash-es'

interface Column {
  key: string
  title: string
  sortable?: boolean
  link?: string
}

interface Props {
  title: string
  columns: Column[]
  total: number
  currentPage: number
  pageSize: number
  loading?: boolean
  showSearch?: boolean
  searchPlaceholder?: string
  pageSizes?: number[]
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

interface Emits {
  (e: 'update:currentPage', page: number): void
  (e: 'update:pageSize', size: number): void
  (e: 'search', query: string): void
  (e: 'sort', key: string): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  showSearch: false,
  searchPlaceholder: '搜索...',
  pageSizes: () => [20, 50, 100],
  sortBy: '',
  sortOrder: 'desc',
})

const emit = defineEmits<Emits>()

const searchQuery = ref('')
const currentPageSize = ref(props.pageSize)

// 计算总页数
const totalPages = computed(() => Math.ceil(props.total / currentPageSize.value))

// 计算显示的页码
const displayedPages = computed(() => {
  const current = props.currentPage
  const total = totalPages.value
  const delta = 2

  let start = Math.max(1, current - delta)
  let end = Math.min(total, current + delta)

  if (end - start < delta * 2) {
    if (start === 1) {
      end = Math.min(start + delta * 2, total)
    } else {
      start = Math.max(end - delta * 2, 1)
    }
  }

  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

// 处理页码变化
const handlePageChange = (page: number) => {
  emit('update:currentPage', page)
}

// 处理每页条数变化
const handleSizeChange = () => {
  emit('update:pageSize', currentPageSize.value)
  handlePageChange(1)
}

// 处理搜索
const handleSearch = debounce(() => {
  emit('search', searchQuery.value)
  handlePageChange(1)
}, 300)

// 处理排序
const handleSort = (key: string) => {
  emit('sort', key)
}

// 监听 pageSize 变化
watch(
  () => props.pageSize,
  (newSize) => {
    currentPageSize.value = newSize
  },
)
</script>

<style scoped>
.list {
  padding: 80px 2rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.list-info {
  display: flex;
  gap: 2rem;
  align-items: center;
}

.search-box input {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-light);
}

.list-table {
  width: 100%;
  border-collapse: collapse;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  overflow: hidden;
}

.list-table th,
.list-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sortable {
  cursor: pointer;
  user-select: none;
}

.sortable:hover {
  background: rgba(255, 255, 255, 0.1);
}

.sort-icon {
  margin-left: 0.5rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-info img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.loading {
  text-align: center;
  padding: 2rem;
}

.pagination {
  margin-top: 2rem;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.pagination-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--text-light);
}

.page-size-select {
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-light);
  cursor: pointer;
}

.pagination-buttons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.page-numbers {
  display: flex;
  gap: 0.5rem;
}

.page-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: var(--primary-color);
  color: white;
}

.page-btn.active {
  background: var(--primary-color);
  color: white;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 适配浅色主题 */
:global(.light-theme) .page-btn {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

:global(.light-theme) .page-size-select {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-dark);
}

@media (max-width: 768px) {
  .list-header {
    flex-direction: column;
    gap: 1rem;
  }

  .list-info {
    flex-direction: column;
    gap: 1rem;
  }

  .list-table {
    font-size: 0.9rem;
  }

  .pagination-buttons {
    flex-wrap: wrap;
    justify-content: center;
  }

  .page-btn {
    padding: 0.4rem 0.8rem;
    font-size: 0.9rem;
  }
}
</style>
