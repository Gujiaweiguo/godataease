<script lang="ts" setup>
import logo from '@/assets/svg/logo.svg'
import aboutBg from '@/assets/img/about-bg.png'
import { reactive, ref, onMounted } from 'vue'
import { useI18n } from '@/hooks/web/useI18n'
import { useUserStoreWithOut } from '@/store/modules/user'
import { buildVersionApi, validateApi } from '@/api/about'
import type { F2CLicense } from './index'

const { t } = useI18n()
const userStore = useUserStoreWithOut()
const build = ref('')
const isAdmin = ref(false)
const tipsSuffix = ref('')
const license: F2CLicense = reactive({
  status: '',
  corporation: '',
  expired: '',
  count: 0,
  version: '',
  edition: '',
  serialNo: '',
  remark: '',
  isv: ''
})

const getLicense = result => {
  if (result.status === 'valid') {
    tipsSuffix.value =
      result?.license?.edition === 'Enterprise' ? t('about.count_of') : t('about.set_of')
  }
  return {
    status: result.status,
    corporation: result.license ? result.license.corporation : '',
    expired: result.license ? result.license.expired : '',
    count: result.license ? result.license.count : 0,
    version: result.license ? result.license.version : '',
    edition: result.license ? result.license.edition : '',
    serialNo: result.license ? result.license.serialNo : '',
    remark: result.license ? result.license.remark : '',
    isv: result.license ? result.license.isv : ''
  }
}

onMounted(async () => {
  isAdmin.value = userStore.getUid === '1'
  const version = await buildVersionApi()
  build.value = version.data
  const result = await validateApi({})
  Object.assign(license, getLicense(result.data))
})
</script>

<template>
  <div class="about-page">
    <div class="about-card">
      <img class="about-bg" :src="aboutBg" />
      <el-icon class="logo"><icon name="logo"><logo class="svg-icon" /></icon></el-icon>
      <div class="content">
        <div class="item"><div class="label">{{ $t('about.auth_to') }}</div><div class="value">{{ license.corporation || '-' }}</div></div>
        <div class="item" v-if="license.isv"><div class="label">ISV</div><div class="value">{{ license.isv }}</div></div>
        <div class="item"><div class="label">{{ $t('about.expiration_time') }}</div><div class="value">{{ license.expired || '-' }}</div></div>
        <div class="item"><div class="label">{{ $t('about.auth_num') }}</div><div class="value">{{ license.status === 'valid' ? `${license.count} ${tipsSuffix}` : '-' }}</div></div>
        <div class="item"><div class="label">{{ $t('about.version_num') }}</div><div class="value">{{ build || '-' }}</div></div>
        <div class="item"><div class="label">{{ $t('about.serial_no') }}</div><div class="value">{{ license.serialNo || '-' }}</div></div>
        <div class="item"><div class="label">{{ $t('about.remark') }}</div><div class="value">{{ license.remark || '-' }}</div></div>
        <div class="item"><div class="label">{{ $t('commons.user_id') }}</div><div class="value">{{ isAdmin ? 'admin' : '-' }}</div></div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.about-page {
  padding: 24px;
  display: flex;
  justify-content: center;
}
.about-card {
  width: 840px;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  position: relative;
}
.about-bg {
  width: 100%;
  height: 180px;
  object-fit: cover;
}
.logo {
  position: absolute;
  top: 28px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 120px;
  color: #fff;
}
.content {
  padding: 24px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 24px;
}
.item .label {
  color: #8d9199;
  font-size: 14px;
}
.item .value {
  color: #1f2329;
  font-size: 16px;
  margin-top: 4px;
  word-break: break-all;
}
</style>
