<script lang="ts" setup>
import noLic from './nolic.vue'
import { ref, shallowRef, useAttrs, onMounted } from 'vue'
import { loadDistributed, xpackModelApi } from '@/api/plugin'
import configGlobal from '@/components/config-global/src/ConfigGlobal.vue'
import { useCache } from '@/hooks/web/useCache'
import { i18n } from '@/plugins/vue-i18n'
import * as Vue from 'vue'
import axios from 'axios'
import * as Pinia from 'pinia'
import * as echarts from 'echarts'
import router from '@/router'
import tinymce from 'tinymce/tinymce'
import { useEmitt } from '@/hooks/web/useEmitt'
import { isNull } from '@/utils/utils'

const { wsCache } = useCache()

const plugin = shallowRef()

const loading = ref(false)

const attrs = useAttrs()

const showNolic = () => {
  plugin.value = noLic
  loading.value = false
}

const loadXpack = async () => {
  if (window['DEXPack']) {
    const xpack = await window['DEXPack'].mapping[attrs.jsname]
    plugin.value = xpack.default
  }
}

useEmitt({
  name: 'load-xpack',
  callback: loadXpack
})

const pluginProxy = ref(null)
const invokeMethod = param => {
  if (pluginProxy.value && pluginProxy.value['invokeMethod']) {
    pluginProxy.value['invokeMethod'](param)
  } else if (param.methodName && pluginProxy.value[param.methodName]) {
    pluginProxy.value[param.methodName](param.args)
  }
}
const emits = defineEmits(['loadFail'])
defineExpose({
  invokeMethod
})
onMounted(async () => {
  const key = 'xpack-model-distributed'
  let distributed: any = false
  if (wsCache.get(key) === null) {
    const res = await xpackModelApi()
    const resData = isNull(res.data) ? 'null' : res.data
    wsCache.set('xpack-model-distributed', resData)
    distributed = res.data
  } else {
    distributed = wsCache.get(key)
  }
  // Normalize wsCache serialization: 'false' string should be boolean false
  if (distributed === 'false' || distributed === 0) {
    distributed = false
  }
  if (isNull(distributed)) {
    setTimeout(() => {
      emits('loadFail')
      loading.value = false
    }, 1000)
    return
  }
  if (distributed) {
    if (window['DEXPack']) {
      const xpack = await window['DEXPack'].mapping[attrs.jsname]
      plugin.value = xpack.default
    } else if (!(window as any)._de_xpack_not_loaded) {
      ;(window as any)._de_xpack_not_loaded = true
      ;(window as any)['VueDe'] = Vue
      ;(window as any)['AxiosDe'] = axios
      ;(window as any)['PiniaDe'] = Pinia
      ;(window as any)['vueRouterDe'] = router
      ;(window as any)['MittAllDe'] = useEmitt().emitter.all
      ;(window as any)['I18nDe'] = i18n
      ;(window as any)['EchartsDE'] = echarts
      if (!(window as any).tinymce) {
        ;(window as any).tinymce = tinymce
      }
      loadDistributed()
        .then(async res => {
          new Function(res.data)()
          useEmitt().emitter.emit('load-xpack')
        })
        .catch(() => {
          emits('loadFail')
          showNolic()
        })
    }
  } else {
    emits('loadFail')
    showNolic()
  }
})
</script>

<template>
  <configGlobal>
    <component
      :key="attrs.jsname"
      ref="pluginProxy"
      :is="plugin"
      v-loading="loading"
      v-bind="attrs"
    ></component>
  </configGlobal>
</template>
