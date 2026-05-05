<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus-secondary'
import type { TaskInfoVO, TaskSaveRequest } from '@/api/datafilling'
import { getTaskInfo, saveTask } from '@/api/datafilling'

interface TaskEditorFormModel {
  name: string
  reciFlagList: number[]
  uidListText: string
  ridListText: string
  fillType: number
  fitType: number
  fitColumn: string
  rateType: number
  rateVal: string
  oneTimeType: number
  startTime: string | number | null
  endTime: string | number | null
  publishRangeTime: number
  publishRangeTimeType: number
  formExtSetting: string
  formFilterSetting: string
}

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    formId: number
    taskId?: number
  }>(),
  {
    modelValue: false,
    formId: 0
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value)
})

const detailLoading = ref(false)
const saveLoading = ref(false)
const activePanels = ref<string[]>(['advanced'])

const createDefaultForm = (): TaskEditorFormModel => ({
  name: '',
  reciFlagList: [],
  uidListText: '',
  ridListText: '',
  fillType: 0,
  fitType: 0,
  fitColumn: '',
  rateType: 0,
  rateVal: '',
  oneTimeType: 0,
  startTime: null,
  endTime: null,
  publishRangeTime: 0,
  publishRangeTimeType: 0,
  formExtSetting: '',
  formFilterSetting: ''
})

const formState = ref<TaskEditorFormModel>(createDefaultForm())

const drawerTitle = computed(() => (props.taskId ? '编辑任务' : '创建任务'))
const showFitColumn = computed(() => formState.value.fitType === 1)
const showRateVal = computed(() => formState.value.rateType > 0)
const showUserRecipients = computed(() => formState.value.reciFlagList.includes(0))
const showRoleRecipients = computed(() => formState.value.reciFlagList.includes(1))

const parseIdList = (value: string) => {
  return Array.from(
    new Set(
      value
        .split(/[\n,，\s]+/)
        .map(item => item.trim())
        .filter(Boolean)
        .map(item => Number(item))
        .filter(item => Number.isFinite(item) && item > 0)
    )
  )
}

const normalizeTimestamp = (value: string | number | null) => {
  if (value === null || value === '') {
    return 0
  }

  const nextValue = Number(value)
  return Number.isFinite(nextValue) ? nextValue : 0
}

const resetForm = () => {
  formState.value = createDefaultForm()
  activePanels.value = ['advanced']
}

const applyTaskDetail = (detail: TaskInfoVO) => {
  formState.value = {
    name: detail.name || '',
    reciFlagList: Array.isArray(detail.reciFlagList) ? detail.reciFlagList : [],
    uidListText: Array.isArray(detail.uidList) ? detail.uidList.join(', ') : '',
    ridListText: Array.isArray(detail.ridList) ? detail.ridList.join(', ') : '',
    fillType: typeof detail.fillType === 'number' ? detail.fillType : 0,
    fitType: typeof detail.fitType === 'number' ? detail.fitType : 0,
    fitColumn: detail.fitColumn || '',
    rateType: typeof detail.rateType === 'number' ? detail.rateType : 0,
    rateVal: detail.rateVal || '',
    oneTimeType: typeof detail.oneTimeType === 'number' ? detail.oneTimeType : 0,
    startTime: detail.startTime > 0 ? String(detail.startTime) : null,
    endTime: detail.endTime > 0 ? String(detail.endTime) : null,
    publishRangeTime: typeof detail.publishRangeTime === 'number' ? detail.publishRangeTime : 0,
    publishRangeTimeType:
      typeof detail.publishRangeTimeType === 'number' ? detail.publishRangeTimeType : 0,
    formExtSetting: detail.formExtSetting || '',
    formFilterSetting: detail.formFilterSetting || ''
  }
}

const loadTaskDetail = async () => {
  if (!props.taskId) {
    resetForm()
    return
  }

  detailLoading.value = true
  try {
    const detail = (await getTaskInfo(props.taskId)) as TaskInfoVO
    applyTaskDetail(detail)
  } catch {
    ElMessage.error('加载任务详情失败')
  } finally {
    detailLoading.value = false
  }
}

const initializeEditor = async () => {
  resetForm()
  if (props.taskId) {
    await loadTaskDetail()
  }
}

const validateForm = () => {
  if (!props.formId) {
    ElMessage.warning('未获取到有效 formId，无法保存任务')
    return false
  }
  if (!formState.value.name.trim()) {
    ElMessage.warning('请输入任务名称')
    return false
  }
  if (showFitColumn.value && !formState.value.fitColumn.trim()) {
    ElMessage.warning('请选择或输入匹配字段')
    return false
  }
  if (showRateVal.value && !formState.value.rateVal.trim()) {
    ElMessage.warning('请输入调度频率值')
    return false
  }
  if (!formState.value.reciFlagList.length) {
    ElMessage.warning('请至少配置一种接收人范围')
    return false
  }
  if (showUserRecipients.value && parseIdList(formState.value.uidListText).length === 0) {
    ElMessage.warning('请填写指定用户 ID')
    return false
  }
  if (showRoleRecipients.value && parseIdList(formState.value.ridListText).length === 0) {
    ElMessage.warning('请填写指定角色 ID')
    return false
  }

  const startTime = normalizeTimestamp(formState.value.startTime)
  const endTime = normalizeTimestamp(formState.value.endTime)
  if (startTime > 0 && endTime > 0 && endTime < startTime) {
    ElMessage.warning('结束时间不能早于开始时间')
    return false
  }

  return true
}

const buildPayload = (): TaskSaveRequest => {
  return {
    id: props.taskId,
    formId: props.formId,
    name: formState.value.name.trim(),
    reciFlagList: formState.value.reciFlagList,
    uidList: parseIdList(formState.value.uidListText),
    ridList: parseIdList(formState.value.ridListText),
    fillType: formState.value.fillType,
    fitType: formState.value.fitType,
    fitColumn: showFitColumn.value ? formState.value.fitColumn.trim() : '',
    rateType: formState.value.rateType,
    rateVal: showRateVal.value ? formState.value.rateVal.trim() : '',
    oneTimeType: formState.value.oneTimeType,
    startTime: normalizeTimestamp(formState.value.startTime),
    endTime: normalizeTimestamp(formState.value.endTime),
    publishRangeTime: Math.max(Number(formState.value.publishRangeTime || 0), 0),
    publishRangeTimeType: formState.value.publishRangeTimeType,
    formExtSetting: formState.value.formExtSetting.trim(),
    formFilterSetting: formState.value.formFilterSetting.trim()
  }
}

const handleSave = async () => {
  if (!validateForm()) {
    return
  }

  saveLoading.value = true
  try {
    await saveTask(buildPayload())
    ElMessage.success(props.taskId ? '任务更新成功' : '任务创建成功')
    emit('saved')
    visible.value = false
  } catch {
    ElMessage.error(props.taskId ? '更新任务失败' : '创建任务失败')
  } finally {
    saveLoading.value = false
  }
}

watch(
  [() => visible.value, () => props.taskId, () => props.formId],
  ([open]) => {
    if (open) {
      void initializeEditor()
      return
    }

    resetForm()
  },
  { immediate: false }
)
</script>

<template>
  <el-drawer v-model="visible" :title="drawerTitle" size="640px" direction="rtl" append-to-body>
    <div class="df-task-editor" v-loading="detailLoading">
      <div class="df-task-editor__intro">
        <div>
          <div class="df-task-editor__intro-title">任务配置</div>
          <div class="df-task-editor__intro-subtitle">配置调度策略、接收范围与高级筛选参数。</div>
        </div>
        <div class="df-task-editor__intro-badge">表单 ID：{{ formId }}</div>
      </div>

      <div class="df-task-editor__body">
        <div class="df-task-editor__section">
          <div class="df-task-editor__section-title">基础信息</div>
          <el-form label-position="top" class="df-task-editor__form">
            <div class="df-task-editor__grid">
              <el-form-item label="任务名称" required>
                <el-input v-model="formState.name" placeholder="请输入任务名称" clearable />
              </el-form-item>

              <el-form-item label="填报类型">
                <el-select v-model="formState.fillType" placeholder="请选择填报类型">
                  <el-option :label="'单人填报'" :value="0" />
                  <el-option :label="'多人填报'" :value="1" />
                </el-select>
              </el-form-item>

              <el-form-item label="匹配方式">
                <el-select v-model="formState.fitType" placeholder="请选择匹配方式">
                  <el-option :label="'不限制'" :value="0" />
                  <el-option :label="'按字段匹配'" :value="1" />
                </el-select>
              </el-form-item>

              <el-form-item v-if="showFitColumn" label="匹配字段" required>
                <el-input v-model="formState.fitColumn" placeholder="请输入匹配字段名" clearable />
              </el-form-item>

              <el-form-item label="调度类型">
                <el-select v-model="formState.rateType" placeholder="请选择调度类型">
                  <el-option :label="'单次'" :value="0" />
                  <el-option :label="'每日'" :value="1" />
                  <el-option :label="'每周'" :value="2" />
                  <el-option :label="'每月'" :value="3" />
                  <el-option :label="'每年'" :value="4" />
                </el-select>
              </el-form-item>

              <el-form-item v-if="showRateVal" label="调度值" required>
                <el-input v-model="formState.rateVal" placeholder="请输入调度表达值" clearable />
              </el-form-item>

              <el-form-item label="单次执行方式">
                <el-select v-model="formState.oneTimeType" placeholder="请选择执行方式">
                  <el-option :label="'立即执行'" :value="0" />
                  <el-option :label="'定时执行'" :value="1" />
                </el-select>
              </el-form-item>

              <el-form-item label="开始时间">
                <el-date-picker
                  v-model="formState.startTime"
                  type="datetime"
                  value-format="x"
                  placeholder="请选择开始时间"
                  style="width: 100%"
                />
              </el-form-item>

              <el-form-item label="结束时间">
                <el-date-picker
                  v-model="formState.endTime"
                  type="datetime"
                  value-format="x"
                  placeholder="请选择结束时间"
                  style="width: 100%"
                />
              </el-form-item>

              <el-form-item label="发布时间范围">
                <el-input-number v-model="formState.publishRangeTime" :min="0" style="width: 100%" />
              </el-form-item>

              <el-form-item label="发布时间单位">
                <el-select v-model="formState.publishRangeTimeType" placeholder="请选择单位">
                  <el-option :label="'分钟'" :value="0" />
                  <el-option :label="'小时'" :value="1" />
                  <el-option :label="'天'" :value="2" />
                </el-select>
              </el-form-item>
            </div>
          </el-form>
        </div>

        <div class="df-task-editor__section">
          <div class="df-task-editor__section-title">接收人配置</div>
          <el-form label-position="top" class="df-task-editor__form">
            <el-form-item label="接收范围">
              <el-checkbox-group v-model="formState.reciFlagList">
                <el-checkbox :label="0">指定用户</el-checkbox>
                <el-checkbox :label="1">指定角色</el-checkbox>
              </el-checkbox-group>
            </el-form-item>

            <div class="df-task-editor__grid">
              <el-form-item v-if="showUserRecipients" label="用户 ID 列表">
                <el-input
                  v-model="formState.uidListText"
                  type="textarea"
                  :autosize="{ minRows: 3, maxRows: 5 }"
                  placeholder="请输入用户 ID，多个值可使用逗号、空格或换行分隔"
                />
              </el-form-item>

              <el-form-item v-if="showRoleRecipients" label="角色 ID 列表">
                <el-input
                  v-model="formState.ridListText"
                  type="textarea"
                  :autosize="{ minRows: 3, maxRows: 5 }"
                  placeholder="请输入角色 ID，多个值可使用逗号、空格或换行分隔"
                />
              </el-form-item>
            </div>
          </el-form>
        </div>

        <div class="df-task-editor__section df-task-editor__section--advanced">
          <el-collapse v-model="activePanels">
            <el-collapse-item title="高级设置" name="advanced">
              <el-form label-position="top" class="df-task-editor__form">
                <el-form-item label="表单扩展设置（JSON）">
                  <el-input
                    v-model="formState.formExtSetting"
                    type="textarea"
                    :autosize="{ minRows: 4, maxRows: 8 }"
                    placeholder="请输入 formExtSetting JSON 字符串"
                  />
                </el-form-item>

                <el-form-item label="表单过滤设置（JSON）">
                  <el-input
                    v-model="formState.formFilterSetting"
                    type="textarea"
                    :autosize="{ minRows: 4, maxRows: 8 }"
                    placeholder="请输入 formFilterSetting JSON 字符串"
                  />
                </el-form-item>
              </el-form>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="df-task-editor__footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saveLoading" @click="handleSave">保存任务</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style lang="less" scoped>
.df-task-editor {
  --df-space-1: 4px;
  --df-space-2: 8px;
  --df-space-3: 12px;
  --df-space-4: 16px;
  --df-space-5: 20px;
  --df-space-6: 24px;
  --df-radius-sm: 4px;
  --df-text-primary: var(--deTextPrimary, #1f2329);
  --df-text-secondary: var(--N600, #646a73);
  --df-border-color: var(--deCardStrokeColor, #e8eaec);
  --df-bg-soft: var(--deTextPrimary5, #f5f6f7);
  --df-color-primary: var(--ed-color-primary, #3370ff);
  --df-color-primary-soft: var(--ed-color-primary-1a, rgba(51, 112, 255, 0.1));

  display: flex;
  height: 100%;
  flex-direction: column;

  &__intro {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--df-space-4);
    border: 1px solid var(--df-border-color);
    border-radius: var(--df-radius-sm);
    background: linear-gradient(135deg, var(--df-color-primary-soft), rgba(255, 255, 255, 0));
    padding: var(--df-space-4) var(--df-space-5);
  }

  &__intro-title {
    color: var(--df-text-primary);
    font-size: 16px;
    font-weight: 500;
    line-height: 24px;
  }

  &__intro-subtitle {
    margin-top: var(--df-space-1);
    color: var(--df-text-secondary);
    font-size: 13px;
    line-height: 20px;
  }

  &__intro-badge {
    color: var(--df-color-primary);
    font-size: 12px;
    line-height: 18px;
    white-space: nowrap;
  }

  &__body {
    flex: 1;
    overflow-y: auto;
    margin-top: var(--df-space-4);
    padding-right: var(--df-space-1);
  }

  &__section {
    border: 1px solid var(--df-border-color);
    border-radius: var(--df-radius-sm);
    background: #fff;
    padding: var(--df-space-5);

    & + & {
      margin-top: var(--df-space-4);
    }
  }

  &__section-title {
    color: var(--df-text-primary);
    font-size: 14px;
    font-weight: 500;
    line-height: 22px;
    margin-bottom: var(--df-space-4);
  }

  &__form {
    width: 100%;
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 var(--df-space-4);
  }

  &__section--advanced {
    background: var(--df-bg-soft);
  }

  &__footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--df-space-3);
  }
}
</style>
