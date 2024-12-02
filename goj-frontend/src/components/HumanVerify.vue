<template>
  <div class="verify-container">
    <div v-if="!verified" class="slider-container" ref="sliderContainer">
      <div class="slider-bg"></div>
      <div class="slider-bar" :style="{ width: sliderValue + '%' }"></div>
      <div
        class="slider-button"
        :class="{ 'slider-button-success': verified }"
        @mousedown="startSlide"
        @touchstart.prevent="startSlide"
        :style="{ left: `calc(${sliderValue}% - 20px)` }"
        ref="sliderButton"
      >
        <span v-if="!sliding">
          <svg viewBox="0 0 24 24" width="16" height="16">
            <path
              fill="currentColor"
              d="M8.59 16.59L13.17 12L8.59 7.41L10 6l6 6l-6 6l-1.41-1.41z"
            />
          </svg>
        </span>
        <span v-else>
          <svg viewBox="0 0 24 24" width="16" height="16">
            <path
              fill="currentColor"
              d="M6.41 6L5 7.41L9.58 12L5 16.59L6.41 18l6-6z M17.59 6l-6 6l6 6L19 16.59L14.42 12L19 7.41z"
            />
          </svg>
        </span>
      </div>
      <span class="slider-text">{{ sliderText }}</span>
    </div>
    <div v-else class="verify-success">
      <svg viewBox="0 0 24 24" width="16" height="16">
        <path fill="currentColor" d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
      </svg>
      验证通过
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  threshold?: number
}>()

const emit = defineEmits<{
  (e: 'verify-success'): void
}>()

const sliderContainer = ref<HTMLElement | null>(null)
const sliderButton = ref<HTMLElement | null>(null)
const sliderValue = ref(0)
const sliding = ref(false)
const verified = ref(false)
const startX = ref(0)

const sliderText = ref('向右滑动验证')

const startSlide = (e: MouseEvent | TouchEvent) => {
  if (verified.value) return

  sliding.value = true
  startX.value = 'touches' in e ? e.touches[0].clientX : e.clientX

  document.addEventListener('mousemove', moving)
  document.addEventListener('mouseup', stopSlide)
  document.addEventListener('touchmove', moving, { passive: false })
  document.addEventListener('touchend', stopSlide)
}

const moving = (e: MouseEvent | TouchEvent) => {
  if (!sliding.value) return
  e.preventDefault()

  const container = sliderContainer.value
  if (!container) return

  const currentX = 'touches' in e ? e.touches[0].clientX : e.clientX
  const containerRect = container.getBoundingClientRect()
  const containerWidth = containerRect.width

  // 计算滑块位置的百分比
  let percent = ((currentX - containerRect.left) / containerWidth) * 100
  percent = Math.max(0, Math.min(100, percent))

  sliderValue.value = percent

  // 如果达到阈值，验证通过
  if (percent > (props.threshold || 90)) {
    verified.value = true
    sliding.value = false
    emit('verify-success')
    removeEvents()
  }
}

const stopSlide = () => {
  if (!verified.value) {
    sliding.value = false
    sliderValue.value = 0
  }
  removeEvents()
}

const removeEvents = () => {
  document.removeEventListener('mousemove', moving)
  document.removeEventListener('mouseup', stopSlide)
  document.removeEventListener('touchmove', moving)
  document.removeEventListener('touchend', stopSlide)
}
</script>

<style scoped>
.verify-container {
  width: 100%;
  height: 40px;
  position: relative;
  margin: 1rem 0;
}

.slider-container {
  width: 100%;
  height: 100%;
  position: relative;
  background: var(--nav-bg-light);
  border-radius: 4px;
  overflow: hidden;
  touch-action: none;
}

.slider-bg {
  position: absolute;
  width: 100%;
  height: 100%;
  background: var(--nav-bg-light);
}

.slider-bar {
  position: absolute;
  height: 100%;
  background: var(--primary-color);
  transition: width 0.1s;
}

.slider-button {
  position: absolute;
  width: 40px;
  height: 100%;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-gray);
  transition: all 0.3s;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  user-select: none;
  touch-action: none;
}

.slider-button:hover {
  background-color: #f5f5f5;
}

.slider-button-success {
  background-color: var(--primary-color) !important;
  color: white !important;
}

.slider-text {
  position: absolute;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-gray);
  user-select: none;
  pointer-events: none;
}

.verify-success {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-color);
  color: white;
  border-radius: 4px;
  gap: 8px;
}

.verify-success svg {
  margin-right: 4px;
}

svg {
  width: 16px;
  height: 16px;
  vertical-align: middle;
}
</style>
