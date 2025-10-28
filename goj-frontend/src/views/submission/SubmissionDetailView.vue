<template>
  <div v-if="loading" class="loading-overlay">
    <div class="loading-spinner"></div>
    <div class="loading-text">加载中...</div>
  </div>

  <div class="submission-detail" v-else-if="submission" :class="{ 'fade-in': submission }">
    <div class="left-panel glass-effect">
      <div class="header">
        <div class="title-section">
          <h1>
            提交详情
            <span class="dot-separator">•</span>
            <span class="submission-id">#{{ submission?.id }}</span>
          </h1>
          <div class="problem-info">
            <router-link
              :to="`/problem/${submission?.problemId}`"
              target="_blank"
              rel="noopener noreferrer"
              class="problem-link"
            >
              <span class="problem-id-badge"> #{{ submission?.problemId }} </span>
              <span class="problem-title">
                {{ submission?.problemTitle }}
              </span>
            </router-link>
          </div>
        </div>

        <div class="meta-info">
          <div class="user-status-section glass-card">
            <div class="user-info">
              <span class="label">提交者</span>
              <router-link :to="`/profile/${submission?.username}`" class="value hover-effect">
                {{ submission?.username }}
              </router-link>
            </div>
            <div class="status-badge" :class="getStatusClass(submission?.status)">
              {{ submission?.status }}
            </div>
          </div>

          <div class="stats-section">
            <div class="stats-grid">
              <!-- 提交时间框 (新增) -->
              <div class="stat-item glass-card submit-time-stat">
                <span class="label">提交时间</span>
                <span class="value submit-time-value">
                  {{ formatTime(submission?.submitTime) }}
                </span>
              </div>
              <!-- 语言 -->
              <div class="stat-item glass-card">
                <span class="label">语言</span>
                <span class="value language-badge" :class="submission?.language.toLowerCase()">
                  {{ getLanguageDisplay(submission?.language || '') }}
                </span>
              </div>
              <!-- 耗时 -->
              <div class="stat-item glass-card">
                <span class="label">耗时</span>
                <span class="value time-badge"> {{ submission?.timeUsed }}ms </span>
              </div>
              <!-- 内存 -->
              <div class="stat-item glass-card">
                <span class="label">内存</span>
                <span class="value memory-badge"> {{ submission?.memoryUsed }}KB </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 如果有错误信息则显示 -->
      <div v-if="submission?.errorInfo" class="error-info glass-card">
        <h3>错误信息</h3>
        <pre>{{ submission.errorInfo }}</pre>
      </div>

      <div class="testcase-list" v-if="submission?.testcasesStatus?.length">
        <div
          v-for="(status, index) in submission.testcasesStatus"
          :key="index"
          class="testcase-item glass-card"
          :class="[status.toLowerCase(), { expanded: expandedTests[index] }]"
        >
          <div class="testcase-header" @click="toggleTestCase(index)">
            <div class="testcase-info">
              <span class="testcase-number">测试点 #{{ index + 1 }}</span>
              <span class="status-badge" :class="getStatusClass(status)">
                {{ status }}
              </span>
              <div class="resource-badges" v-if="submission.testCaseResults?.[index]">
                <span class="time-badge">
                  {{ submission.testCaseResults[index].timeUsed }}ms
                </span>
                <span class="memory-badge">
                  {{ submission.testCaseResults[index].memoryUsed }}KB
                </span>
              </div>
            </div>
            <i class="fas fa-chevron-down" :class="{ rotate: expandedTests[index] }"></i>
          </div>
          <div class="testcase-detail" v-if="expandedTests[index]">
            <div class="detail-grid">
              <div class="detail-item">
                {{ submission?.testcasesInfo?.[index] || '无详细信息' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="right-panel glass-effect">
      <div class="code-header">
        <h2>提交的代码</h2>
        <button class="copy-btn" @click="copyCode" :class="{ copied: isCopied }">
          <i class="fas fa-copy"></i>
          {{ isCopied ? '已复制' : '复制代码' }}
        </button>
      </div>
      <div class="code-container">
        <pre
          class="code-block"
        ><code :class="submission?.language.toLowerCase()" v-html="highlightedCode"></code></pre>
      </div>
    </div>
  </div>

  <div v-else class="error-state">
    <i class="fas fa-exclamation-circle"></i>
    <p>加载失败，请稍后重试</p>
  </div>
</template>

<style scoped>
.submission-detail {
  flex: 1;
  display: flex;
  gap: 2rem;
  padding: 0 2rem;
  color: var(--text-primary);
  height: calc(100vh - 80px);
  margin-top: 80px;
}

.glass-effect {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.glass-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(5px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.glass-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
  border-color: var(--primary-color);
}

.left-panel {
  width: 60%;
  padding: 2rem;
  overflow-y: auto;
}

.right-panel {
  width: 40%;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.header {
  margin-bottom: 2rem;
}

.title-section {
  margin-bottom: 1.5rem;
}

.title-section h1 {
  font-size: 1.8rem;
  margin-bottom: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.8rem;
  color: var(--text-primary);
}

.dot-separator {
  color: var(--text-secondary);
  opacity: 0.5;
  font-weight: 300;
  margin: 0 0.2rem;
}

.submission-id {
  font-size: 1.2rem;
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  font-weight: 500;
}

.submission-id:hover {
  transform: translateY(-2px);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.08));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.meta-info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.user-status-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  margin-bottom: 1rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stats-section {
  width: 100%;
}

/* 调整 stats-grid 以适应四个项目 */
.stats-grid {
  display: flex;
  gap: 1rem;
}

.stat-item {
  flex: 1;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.8rem;
}

.stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.stat-item .label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  opacity: 0.8;
}

.stat-item .value {
  font-size: 1rem;
  font-weight: 500;
  color: var(--text-primary);
}

.info-grid {
  gap: 1rem;
  margin-bottom: 2rem;
}

.info-item {
  padding: 1rem;
  transition: all 0.3s ease;
}

.hover-effect {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 500;
  position: relative;
  padding-bottom: 2px;
}

.hover-effect::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 2px;
  background: var(--primary-color);
  transform: scaleX(0);
  transform-origin: right;
  transition: transform 0.3s ease;
}

.hover-effect:hover::after {
  transform: scaleX(1);
  transform-origin: left;
}

.testcase-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  padding: 0.5rem;
  margin-top: 2rem;
  display: flex;
  flex-direction: column;
}

.testcase-item {
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.testcase-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--primary-color), var(--success-color));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.testcase-item:hover::before {
  opacity: 1;
}

.testcase-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.testcase-header:hover {
  background: rgba(255, 255, 255, 0.05);
}

.testcase-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.testcase-number {
  font-weight: 500;
  color: var(--text-secondary);
}

.testcase-status {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  background: linear-gradient(135deg, var(--primary-color), var(--success-color));
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.fa-chevron-down {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  transition: all 0.3s ease;
}

.fa-chevron-down.rotate {
  transform: rotate(180deg);
  background: var(--primary-color);
  color: white;
}

.testcase-detail {
  margin-top: 1rem;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  transform-origin: top;
  animation: slideDown 0.3s ease;
}

.testcase-status {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  background: linear-gradient(135deg, var(--primary-color), var(--success-color));
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

@keyframes slideDown {
  from {
    transform: scaleY(0);
    opacity: 0;
  }
  to {
    transform: scaleY(1);
    opacity: 1;
  }
}

.testcase-item.expanded {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  border-color: var(--primary-color);
}

.detail-grid {
  background: rgba(255, 255, 255, 0.03);
  padding: 0.8rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.detail-item {
  padding: 0.5rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 6px;
}

.code-container {
  flex: 1;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: 8px;
  overflow: auto;
}

.code-block {
  margin: 0;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  font-family: 'Fira Code', monospace;
  font-size: 14px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
}

.code-container::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.code-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.code-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.code-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

:deep(.hljs) {
  background: transparent;
  color: var(--text-primary);
}

.code-header {
  padding: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  height: 60px;
}

.copy-btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  background: var(--primary-color);
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
}

.copy-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.copy-btn.copied {
  background: var(--success-color);
}

.status-badge {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

/* Accepted - 通过 */
.status-badge.accepted {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.3);
}

/* Wrong Answer - 答案错误 */
.status-badge.wrong-answer {
  background: linear-gradient(135deg, #ff416c, #ff4b2b);
  box-shadow: 0 2px 8px rgba(255, 65, 108, 0.3);
}

/* Runtime Error - 运行时错误 */
.status-badge.runtime-error {
  background: linear-gradient(135deg, #f7b733, #fc4a1a);
  box-shadow: 0 2px 8px rgba(247, 183, 51, 0.3);
}

/* Time Limit Exceeded - 超时 */
.status-badge.time-limit-exceeded {
  background: linear-gradient(135deg, #834d9b, #d04ed6);
  box-shadow: 0 2px 8px rgba(131, 77, 155, 0.3);
}

/* Memory Limit Exceeded - 内存超限 */
.status-badge.memory-limit-exceeded {
  background: linear-gradient(135deg, #4568dc, #b06ab3);
  box-shadow: 0 2px 8px rgba(69, 104, 220, 0.3);
}

/* Compilation Error - 编译错误 */
.status-badge.compile-error {
  background: linear-gradient(135deg, #373b44, #4286f4);
  box-shadow: 0 2px 8px rgba(55, 59, 68, 0.3);
}

/* System Error - 系统错误 */
.status-badge.system-error {
  background: linear-gradient(135deg, #cb356b, #bd3f32);
  box-shadow: 0 2px 8px rgba(203, 53, 107, 0.3);
}

/* Pending - 等待中 */
.status-badge.pending {
  background: linear-gradient(135deg, #2c3e50, #3498db);
  box-shadow: 0 2px 8px rgba(44, 62, 80, 0.3);
  animation: pulse 2s infinite;
}

/* Judging - 评测中 */
.status-badge.judging {
  background: linear-gradient(135deg, #2c3e50, #3498db);
  box-shadow: 0 2px 8px rgba(44, 62, 80, 0.3);
  animation: pulse 2s infinite;
}

/* 添加脉冲动画 */
@keyframes pulse {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

/* 错误信息样式 */
.error-info {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 8px;
  background: rgba(255, 65, 108, 0.1);
  border: 1px solid rgba(255, 65, 108, 0.2);
}

.error-info h3 {
  color: #ff416c;
  margin-bottom: 0.5rem;
}

.error-info pre {
  margin: 0;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Fira Code', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
}

.fade-in {
  animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.status-badge,
.testcase-status {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
}

/* 通过状态 */
.status-badge.accepted,
.testcase-status.accepted {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  color: white;
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.3);
}

/* 未通过状态 */
.status-badge.rejected,
.testcase-status.rejected {
  background: linear-gradient(135deg, #ff416c, #ff4b2b);
  color: white;
  box-shadow: 0 2px 8px rgba(255, 65, 108, 0.3);
}

/* 测试用例卡片顶部渐变条对应状态 */
.testcase-item.accepted::before {
  background: linear-gradient(90deg, #00b09b, #96c93d);
}

.testcase-item.rejected::before {
  background: linear-gradient(90deg, #ff416c, #ff4b2b);
}

/* 展开箭头对应状态颜色 */
.testcase-item.accepted .fa-chevron-down.rotate {
  background: #00b09b;
}

.testcase-item.rejected .fa-chevron-down.rotate {
  background: #ff416c;
}

/* 确保代码块内的代码不会有多余的缩进 */
.code-block code {
  display: block;
  padding: 0;
  margin: 0;
  white-space: pre;
}

/* 通用徽章样式 */
.value.language-badge,
.value.time-badge,
.value.memory-badge,
.value.submit-time-value /* 新增 */ {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  color: white;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

/* 提交时间徽章样式 (新增) */
.value.submit-time-value {
  background: linear-gradient(135deg, #74ebd5, #9face6); /* 浅蓝色调渐变 */
  color: #333; /* 浅色背景配深色文字 */
  box-shadow: 0 2px 8px rgba(116, 235, 213, 0.3);
}

/* 语言徽章样式 */
.language-badge.cpp {
  background: linear-gradient(135deg, #0072c6, #00a4db);
  box-shadow: 0 2px 8px rgba(0, 114, 198, 0.3);
}

.language-badge.c {
  background: linear-gradient(135deg, #283593, #5c6bc0);
  box-shadow: 0 2px 8px rgba(40, 53, 147, 0.3);
}

.language-badge.java {
  background: linear-gradient(135deg, #f44336, #ff7043);
  box-shadow: 0 2px 8px rgba(244, 67, 54, 0.3);
}

.language-badge.python {
  background: linear-gradient(135deg, #ffd54f, #ffb300);
  box-shadow: 0 2px 8px rgba(255, 213, 79, 0.3);
}

.language-badge.go {
  background: linear-gradient(135deg, #00bcd4, #26c6da);
  box-shadow: 0 2px 8px rgba(0, 188, 212, 0.3);
}

.language-badge.javascript {
  background: linear-gradient(135deg, #f7df1e, #ffd700);
  box-shadow: 0 2px 8px rgba(247, 223, 30, 0.3);
  color: #333; /* JavaScript 的黄色背景配深色文字 */
}

.language-badge.typescript {
  background: linear-gradient(135deg, #3178c6, #5c9aef);
  box-shadow: 0 2px 8px rgba(49, 120, 198, 0.3);
}

/* 耗时徽章样式 */
.time-badge {
  background: linear-gradient(135deg, #4caf50, #81c784);
  box-shadow: 0 2px 8px rgba(76, 175, 80, 0.3);
}

/* 内存徽章样式 */
.memory-badge {
  background: linear-gradient(135deg, #9c27b0, #ba68c8);
  box-shadow: 0 2px 8px rgba(156, 39, 176, 0.3);
}

/* 添加悬停效果 */
.value.language-badge,
.value.time-badge,
.value.memory-badge,
.value.submit-time-value /* 新增 */ {
  transition: all 0.3s ease;
}

.value.language-badge:hover,
.value.time-badge:hover,
.value.memory-badge:hover,
.value.submit-time-value:hover /* 新增 */ {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.problem-info {
  margin-top: 0.5rem;
}

.problem-link {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  text-decoration: none;
  transition: all 0.3s ease;
}

.problem-id-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: var(--text-primary);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.05));
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.problem-title {
  font-size: 1rem;
  color: var(--text-primary);
  opacity: 0.9;
  transition: all 0.3s ease;
  position: relative;
}

.problem-title::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 100%;
  height: 1px;
  background: var(--primary-color);
  transform: scaleX(0);
  transform-origin: right;
  transition: transform 0.3s ease;
}

/* 悬停效果 */
.problem-link:hover .problem-id-badge {
  transform: translateY(-2px);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15), rgba(255, 255, 255, 0.08));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.problem-link:hover .problem-title {
  color: var(--primary-color);
}

.problem-link:hover .problem-title::after {
  transform: scaleX(1);
  transform-origin: left;
}

/* 添加 loading 样式 */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s linear infinite;
}

.loading-text {
  margin-top: 1rem;
  color: var(--text-light);
  font-size: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 添加错误状态样式 */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 80px);
  margin-top: 80px;
  color: var(--text-secondary);
}

.error-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: var(--error-color);
}

.error-state p {
  font-size: 1.2rem;
}

/* 修改测试点状态徽章样式，与主状态徽章保持一致 */
.testcase-info .status-badge {
  padding: 0.4rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
}

/* 状态徽章样式 */
.status-badge.accepted {
  background: linear-gradient(135deg, #00b09b, #96c93d);
  box-shadow: 0 2px 8px rgba(0, 176, 155, 0.3);
}

.status-badge.wrong-answer {
  background: linear-gradient(135deg, #ff416c, #ff4b2b);
  box-shadow: 0 2px 8px rgba(255, 65, 108, 0.3);
}

.status-badge.runtime-error {
  background: linear-gradient(135deg, #f7b733, #fc4a1a);
  box-shadow: 0 2px 8px rgba(247, 183, 51, 0.3);
}

.status-badge.time-limit-exceeded {
  background: linear-gradient(135deg, #834d9b, #d04ed6);
  box-shadow: 0 2px 8px rgba(131, 77, 155, 0.3);
}

.status-badge.memory-limit-exceeded {
  background: linear-gradient(135deg, #4568dc, #b06ab3);
  box-shadow: 0 2px 8px rgba(69, 104, 220, 0.3);
}

.status-badge.compile-error {
  background: linear-gradient(135deg, #373b44, #4286f4);
  box-shadow: 0 2px 8px rgba(55, 59, 68, 0.3);
}

.status-badge.pending,
.status-badge.judging {
  background: linear-gradient(135deg, #2c3e50, #3498db);
  box-shadow: 0 2px 8px rgba(44, 62, 80, 0.3);
  animation: pulse 2s infinite;
}

.status-badge.nonzero-exit-status {
  background: linear-gradient(135deg, #cb356b, #bd3f32);
  box-shadow: 0 2px 8px rgba(203, 53, 107, 0.3);
}

/* 悬停效果 */
.status-badge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* 脉冲动画 */
@keyframes pulse {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

.resource-badges {
  display: flex;
  gap: 0.8rem;
  margin-left: 1rem;
}

/* 时间徽章样式 */
.time-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: white;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  box-shadow: 0 2px 8px rgba(79, 172, 254, 0.3);
  transition: all 0.3s ease;
}

/* 内存徽章样式 */
.memory-badge {
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-block;
  text-align: center;
  color: white;
  background: linear-gradient(135deg, #834d9b, #d04ed6);
  box-shadow: 0 2px 8px rgba(131, 77, 155, 0.3);
  transition: all 0.3s ease;
}

/* 悬停效果 */
.time-badge:hover,
.memory-badge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* 调整测试点布局 */
.testcase-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.testcase-header:hover {
  background: rgba(255, 255, 255, 0.05);
}

/* 确保徽章在一行显示 */
@media (max-width: 768px) {
  .testcase-info {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .resource-badges {
    width: 100%;
    margin-left: 0;
    margin-top: 0.5rem;
    justify-content: flex-start;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/modules/app'
import hljs from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()

// 添加加载状态
const loading = ref(true)

interface TestCaseResult {
  status: string
  timeUsed: number
  memoryUsed: number
  errorInfo: string
}

interface Submission {
  id: number
  userId: number
  username: string
  status: string
  language: string
  timeUsed: number
  memoryUsed: number
  code: string
  errorInfo: string
  problemTitle: string
  problemId: string
  testcasesStatus: string[] | null
  testcasesInfo: string[] | null
  submitTime: string
  judgeTime: string
  testCaseResults?: TestCaseResult[]
}

const submission = ref<Submission | null>(null)
const isCopied = ref(false)
const expandedTests = ref<boolean[]>([])

// 格式化时间函数 (新增)
const formatTime = (timeString: string | undefined): string => {
  if (!timeString) return 'N/A'
  try {
    const date = new Date(timeString)
    // 格式化为：YYYY-MM-DD HH:mm:ss
    const year = date.getFullYear()
    // getMonth() 返回 0-11，所以需要 +1
    const month = String(date.getMonth() + 1).padStart(2, '0') 
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    const seconds = String(date.getSeconds()).padStart(2, '0')

    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
  } catch (e) {
    console.error('时间格式化错误:', e)
    return timeString.substring(0, 19) // 尝试返回 ISO 格式的简化版本
  }
}


// 修改计算属性，添加空值检查
const highlightedCode = computed(() => {
  if (!submission.value?.code || !submission.value?.language) return ''
  const language = submission.value.language.toLowerCase()
  try {
    return hljs.highlight(submission.value.code, {
      language: language,
    }).value
  } catch {
    return submission.value.code
  }
})

// 修改语言显示函数，添加空值检查
const getLanguageDisplay = (lang: string = '') => {
  const langMap: { [key: string]: string } = {
    cpp: 'C++',
    c: 'C',
    java: 'Java',
    python: 'Python',
    go: 'Go',
    javascript: 'JavaScript',
    typescript: 'TypeScript',
  }
  return langMap[lang.toLowerCase()] || lang
}

// 修改获取提交详情的函数，添加数据转换
const fetchSubmissionDetail = async () => {
  loading.value = true
  try {
    const response = await fetch(`/api/submission/${route.params.id}`, {
      headers: {
        Authorization: `Bearer ${userStore.token}`,
      },
    })

    if (!response.ok) {
      throw new Error('获取提交详情失败')
    }

    const data = await response.json()
    console.log('API Response:', data)

    if (data.code === 200) {
      submission.value = {
        id: data.data.id,
        userId: data.data.userId,
        username: data.data.username,
        status: data.data.status,
        language: data.data.language,
        timeUsed: data.data.timeUsed,
        memoryUsed: data.data.memoryUsed,
        code: data.data.code,
        errorInfo: data.data.errorInfo,
        problemTitle: data.data.problemTitle,
        problemId: data.data.problemId,
        testcasesStatus: data.data.testcasesStatus || [],
        testcasesInfo: data.data.testcasesInfo || [],
        submitTime: data.data.submitTime, // 确保这个字段从 API 响应中获取
        judgeTime: data.data.judgeTime,
        testCaseResults: data.data.testCaseResults || [],
      }

      if (submission.value && submission.value.testcasesStatus) {
        expandedTests.value = new Array(submission.value.testcasesStatus.length).fill(false)
      }
    } else {
      throw new Error(data.message)
    }
  } catch (error) {
    console.error('获取提交详情失败:', error)
    appStore.showNotification('error', '获取提交详情失败')
  } finally {
    loading.value = false
  }
}

// 添加复制函数
const copyCode = () => {
  if (!submission.value?.code) return

  // 创建临时文本区域
  const textArea = document.createElement('textarea')
  textArea.value = submission.value.code

  // 设置样式使其不可见
  textArea.style.position = 'fixed'
  textArea.style.left = '-9999px'
  textArea.style.top = '0'

  document.body.appendChild(textArea)

  try {
    // 选择文本
    textArea.select()
    textArea.setSelectionRange(0, 99999) // 对于移动设备

    // 执行复制命令
    const successful = document.execCommand('copy')

    if (successful) {
      isCopied.value = true
      appStore.showNotification('success', '复制成功')
    } else {
      throw new Error('复制失败')
    }
  } catch (err) {
    appStore.showNotification('error', '复制失败，请手动复制')
    console.error('复制失败:', err)
  } finally {
    // 清理
    document.body.removeChild(textArea)

    // 重置复制状态
    setTimeout(() => {
      isCopied.value = false
    }, 2000)
  }
}

// 格式化时间

onMounted(() => {
  fetchSubmissionDetail()
})

onUnmounted(() => {
  // 移除 monaco 相关的码和变量
})

// 添加切换测试点展开状态的函数
const toggleTestCase = (index: number) => {
  expandedTests.value[index] = !expandedTests.value[index]
}

const getStatusClass = (status: string) => {
  // 将状态转换为小写并替换空格为连字符
  return status.toLowerCase().replace(/\s+/g, '-')
}
</script>
