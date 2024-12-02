<template>
  <div class="groups">
    <div class="groups-header">
      <h1>我的群组</h1>
      <button class="create-btn">创建群组</button>
    </div>

    <div class="groups-grid">
      <div class="group-card" v-for="i in 6" :key="i">
        <div class="group-header">
          <div class="group-info">
            <h3>{{ i % 2 === 0 ? '算法学习小组' : 'ACM集训队' }} {{ i }}</h3>
            <span class="member-count">{{ 20 + i * 5 }} 名成员</span>
          </div>
          <div class="group-type" :class="i % 2 === 0 ? 'public' : 'private'">
            {{ i % 2 === 0 ? '公开' : '私有' }}
          </div>
        </div>

        <div class="group-stats">
          <div class="stat">
            <div class="stat-value">{{ 50 + i * 10 }}</div>
            <div class="stat-label">题目</div>
          </div>
          <div class="stat">
            <div class="stat-value">{{ 5 + i }}</div>
            <div class="stat-label">比赛</div>
          </div>
          <div class="stat">
            <div class="stat-value">{{ 100 + i * 20 }}</div>
            <div class="stat-label">讨论</div>
          </div>
        </div>

        <div class="group-actions">
          <button class="action-btn primary">进入群组</button>
          <button class="action-btn">群组设置</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'GroupsView',
})
</script>

<style scoped>
.groups {
  min-height: calc((100vh - 64px - 150px) * 1.5);
  padding: 80px 2rem 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.groups-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.create-btn {
  background: var(--primary-color);
  color: var(--bg-dark);
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
}

.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 2rem;
}

.group-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 1.5rem;
  transition: transform 0.3s;
}

.group-card:hover {
  transform: translateY(-5px);
}

.group-header {
  position: relative;
  margin-bottom: 1.5rem;
}

.group-info h3 {
  margin: 0 0 0.25rem 0;
  color: var(--text-light);
}

.member-count {
  font-size: 0.875rem;
  color: var(--text-gray);
}

.group-type {
  position: absolute;
  top: 0;
  right: 0;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
}

.group-type.public {
  background: rgba(66, 185, 131, 0.2);
  color: #00ADD8;
}

.group-type.private {
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.group-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  padding: 1rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 1.5rem;
}

.stat {
  text-align: center;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: bold;
  color: var(--primary-color);
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-gray);
  margin-top: 0.25rem;
}

.group-actions {
  display: flex;
  gap: 1rem;
}

.action-btn {
  flex: 1;
  padding: 0.5rem;
  border-radius: 4px;
  border: 1px solid var(--text-gray);
  background: transparent;
  color: var(--text-light);
  cursor: pointer;
  transition: all 0.3s;
}

.action-btn.primary {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: var(--bg-dark);
}

.action-btn:hover {
  opacity: 0.9;
}
</style>
