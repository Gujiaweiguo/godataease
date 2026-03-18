<script lang="ts" setup>
import noLic from './nolic.vue'
import { ref, useAttrs, onMounted } from 'vue'
import { xpackModelApi } from '@/api/plugin'
import { useCache } from '@/hooks/web/useCache'
import { i18n } from '@/plugins/vue-i18n'
import * as Vue from 'vue'
import axios from 'axios'
import * as Pinia from 'pinia'
import router from '@/router'
import { useEmitt } from '@/hooks/web/useEmitt'
import request from '@/config/axios'
const { wsCache } = useCache()
import { isNull } from '@/utils/utils'

const plugin = ref()

const loading = ref(false)

const attrs = useAttrs()

const showNolic = () => {
  plugin.value = noLic
  loading.value = false
}

const getModuleName = () => {
  const jsPath = window.atob(attrs.jsname.toString())
  return jsPath.split('/')[0]
}
const pluginProxy = ref(null)
const invokeMethod = param => {
  if (pluginProxy.value['invokeMethod']) {
    pluginProxy.value['invokeMethod'](param)
  } else {
    pluginProxy.value[param.methodName]?.(...param.args)
  }
}

onMounted(async () => {
  const key = 'xpack-model-distributed'
  let distributed = false
  if (wsCache.get(key) === null) {
    const res = await xpackModelApi()
    wsCache.set('xpack-model-distributed', isNull(res.data) ? 'null' : res.data)
    distributed = res.data
  } else {
    distributed = wsCache.get(key)
  }
  if (isNull(distributed)) {
    setTimeout(() => {
      emits('loadFail')
      loading.value = false
    }, 1000)
    return
  }
  if (distributed) {
    const moduleName = getModuleName()
    if (window[moduleName]) {
      const xpack = await window[moduleName].mapping[attrs.jsname]
      plugin.value = xpack.default
    } else {
      window['VueDe'] = Vue
      window['AxiosDe'] = axios
      window['PiniaDe'] = Pinia
      window['vueRouterDe'] = router
      window['MittAllDe'] = useEmitt().emitter.all
      window['I18nDe'] = i18n
      const url = `/xpackComponent/pluginStaticInfo/${moduleName}`
      request.get({ url }).then(async res => {
        new Function(res.data || res)()
        const xpack = await window[moduleName].mapping[attrs.jsname]
        plugin.value = xpack.default
      })
    }
  } else {
    emits('loadFail')
    showNolic()
  }
})

const emits = defineEmits(['loadFail'])
defineExpose({
  invokeMethod
})
</script>

<template>
  <component
    :key="attrs.jsname"
    ref="pluginProxy"
    :is="plugin"
    v-loading="loading"
    v-bind="attrs"
  ></component>
</template>

<style lang="less" scoped></style>
