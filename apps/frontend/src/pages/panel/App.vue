<script setup lang="ts">
import { shallowRef, defineAsyncComponent, ref, onMounted, nextTick } from 'vue'
import { propTypes } from '@/utils/propTypes'
import { useEmitt } from '@/hooks/web/useEmitt'
import { useRouter } from 'vue-router_2'

const VisualizationEditor = defineAsyncComponent(
  () => import('@/views/data-visualization/index.vue')
)
const DashboardEditor = defineAsyncComponent(() => import('@/views/dashboard/index.vue'))

const Dashboard = defineAsyncComponent(() => import('./DashboardPreview.vue'))
const ViewWrapper = defineAsyncComponent(() => import('./ViewWrapper.vue'))
const Iframe = defineAsyncComponent(() => import('./Iframe.vue'))
const Dataset = defineAsyncComponent(() => import('@/views/visualized/data/dataset/index.vue'))
const DatasetEditor = defineAsyncComponent(
  () => import('@/views/visualized/data/dataset/form/index.vue')
)
const Datasource = defineAsyncComponent(
  () => import('@/views/visualized/data/datasource/index.vue')
)
const ScreenPanel = defineAsyncComponent(() => import('@/views/data-visualization/PreviewShow.vue'))
const DashboardPanel = defineAsyncComponent(
  () => import('@/views/dashboard/DashboardPreviewShow.vue')
)

const TemplateManage = defineAsyncComponent(() => import('@/views/template/indexInject.vue'))

const Preview = defineAsyncComponent(() => import('@/views/data-visualization/PreviewCanvas.vue'))
const DashboardEmpty = defineAsyncComponent(() => import('@/views/mobile/panel/DashboardEmpty.vue'))

const props = defineProps({
  componentName: propTypes.string.def('Iframe')
})
const currentComponent = shallowRef()

const componentMap = {
  DashboardEditor,
  VisualizationEditor,
  ViewWrapper,
  Preview,
  Dashboard,
  Dataset,
  Iframe,
  Datasource,
  ScreenPanel,
  DashboardPanel,
  DatasetEditor,
  DashboardEmpty,
  TemplateManage
}

const showComponent = ref(false)
const router = useRouter()

const changeCurrentComponent = val => {
  showComponent.value = true
  currentComponent.value = undefined
  if (val && val.includes('DataFilling')) {
    if (val === 'DataFilling') {
      router.push('/data-filling-manage')
    } else if (val === 'DataFillingEditor') {
      router.push({ path: '/data-filling-editor' })
    } else if (val === 'DataFillingHandler') {
      router.push('/data-filling-fill')
    }
    return
  }
  nextTick(() => {
    currentComponent.value = componentMap[val]
    showComponent.value = false
  })
}

useEmitt({
  name: 'changeCurrentComponent',
  callback: changeCurrentComponent
})

onMounted(() => {
  changeCurrentComponent(props.componentName)
})
</script>
<template>
  <component :is="currentComponent" v-if="currentComponent"></component>
</template>
