<template>
  <div class="admin-layout">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="logo">
        <el-icon><Monitor /></el-icon>
        <span>GOJ Admin</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        background-color="#001529"
        text-color="#fff"
        router
      >
        <!-- 首页 -->
        <el-menu-item index="/admin">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>

        <el-sub-menu index="problem">
          <template #title>
            <el-icon><Grid /></el-icon>
            <span>题目管理</span>
          </template>
          <el-menu-item index="/admin/problem/add">添加题目</el-menu-item>
          <el-menu-item index="/admin/problem/manage">管理题目</el-menu-item>
          <el-menu-item index="/admin/problem/import-export">导入导出</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="contest">
          <template #title>
            <el-icon><Calendar /></el-icon>
            <span>比赛管理</span>
          </template>
          <el-menu-item index="/admin/contest/add">添加比赛</el-menu-item>
          <el-menu-item index="/admin/contest/manage">管理比赛</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="management">
          <template #title>
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/admin/users">用户管理</el-menu-item>
          <!-- <el-menu-item index="/admin/roles">角色管理</el-menu-item> -->
        </el-sub-menu>

        <el-sub-menu index="website">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>网站设置</span>
          </template>
          <el-menu-item index="/admin/website/basic">基本设置</el-menu-item>
        </el-sub-menu>

        <el-menu-item @click="handleSystemRestore">
          <el-icon><RefreshLeft /></el-icon>
          <span>系统还原</span>
        </el-menu-item>

        <!-- <el-menu-item index="project-config">
          <el-icon><Tools /></el-icon>
          <span>项目配置</span>
        </el-menu-item> -->
      </el-menu>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <router-view></router-view>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { Monitor, Setting, Grid, HomeFilled, RefreshLeft } from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/modules/user'

// 组件名称
defineOptions({
  name: 'AdminLayout',
})

const route = useRoute()
const activeMenu = ref(route.path)

const userStore = useUserStore()

const handleSystemRestore = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要还原系统吗？此操作将清空所有数据，包括用户、题目、提交记录等，且不可恢复！',
      '警告',
      {
        confirmButtonText: '确定还原',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger',
      },
    )

    const response = await fetch('/api/admin/system/restore', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('系统还原失败')
    }

    const result = await response.json()
    if (result.code === 200) {
      ElMessage.success('系统已还原')
      // 可选：重新加载页面
      window.location.reload()
    } else {
      throw new Error(result.message || '系统还原失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : '系统还原失败')
    }
  }
}
</script>

<style scoped>
.admin-layout {
  position: fixed;
  top: 64px; /* 顶部导航栏高度 */
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  overflow: hidden; /* 防止整体滚动 */
}

.sidebar {
  width: 240px;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  color: var(--text-light);
  flex-shrink: 0;
  overflow-y: auto; /* 允许侧边栏滚动 */
}

.logo {
  height: 64px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.05);
}

.logo .el-icon {
  font-size: 24px;
}

.sidebar-menu {
  border-right: none;
  background-color: transparent !important;
}

.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  color: var(--text-light);
  opacity: 0.8;
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  color: var(--text-light);
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  color: var(--primary-color);
  background: rgba(var(--primary-color-rgb), 0.1);
}

.main-content {
  flex: 1;
  background: var(--nav-bg-dark);
  backdrop-filter: blur(10px);
  overflow-y: auto; /* 允许内容区滚动 */
  padding: 20px;
}

.new-tag {
  margin-left: 8px;
  background-color: var(--primary-color);
  border: none;
}

/* 图标样式 */
.el-icon {
  vertical-align: middle;
  margin-right: 8px;
  width: 24px;
  text-align: center;
}

/* 菜单项文字样式 */
.el-menu-item span,
.el-sub-menu__title span {
  margin-left: 4px;
}

/* 子菜单样式 */
.sidebar-menu :deep(.el-menu--inline) {
  background: rgba(255, 255, 255, 0.05) !important;
}

.sidebar-menu :deep(.el-sub-menu.is-active .el-sub-menu__title) {
  color: var(--primary-color);
}
</style>
