import SockJS from 'sockjs-client/dist/sockjs.min.js'
import Stomp from 'stompjs'
import { useCache } from '@/hooks/web/useCache'
import { useEmitt } from '@/hooks/web/useEmitt'
const { wsCache } = useCache()
let stompClient: Stomp.Client
let websocketEnabled: boolean | null = null
import dev from '../../config/dev'
const env = import.meta.env
const basePath = env.VITE_API_BASEPATH

const resolveWebSocketEnabled = async (prefix: string): Promise<boolean> => {
  if (websocketEnabled !== null) {
    return websocketEnabled
  }
  try {
    const response = await fetch(prefix + 'websocket/info', {
      method: 'GET',
      credentials: 'include'
    })
    if (!response.ok) {
      websocketEnabled = false
      return websocketEnabled
    }
    const info = await response.json()
    websocketEnabled = info?.websocket === true
    return websocketEnabled
  } catch (error) {
    console.error('获取 websocket 状态失败:', error)
    websocketEnabled = false
    return websocketEnabled
  }
}

export default {
  install() {
    const channels = [
      {
        topic: '/task-export-topic',
        event: 'task-export-topic-call'
      },
      {
        topic: '/report-notice',
        event: 'report-notice-call'
      }
    ]
    function isLoginStatus() {
      if (wsCache.get('app.desktop')) {
        return true
      }
      return wsCache.get('user.token') && wsCache.get('user.uid')
    }

    async function connection() {
      if (!isLoginStatus()) {
        return
      }
      if (stompClient && stompClient.connected) {
        return
      }
      let prefix = '/'
      if (window.DataEaseBi?.baseUrl) {
        prefix = window.DataEaseBi.baseUrl
      } else {
        // const href = window.location.href
        prefix = location.origin + location.pathname
        if (env.MODE === 'dev') {
          prefix = dev.server.proxy[basePath].target + '/'
        }
      }
      if (!prefix.endsWith('/')) {
        prefix += '/'
      }
      const enabled = await resolveWebSocketEnabled(prefix)
      if (!enabled) {
        disconnect()
        return
      }
      const userId = wsCache.get('app.desktop') ? 1 : wsCache.get('user.uid')
      const socket = new SockJS(prefix + 'websocket?userId=' + userId)
      stompClient = Stomp.over(socket)
      const heads = {
        userId: userId
      }
      stompClient.connect(
        heads,
        () => {
          channels.forEach(channel => {
            stompClient.subscribe('/user/' + userId + channel.topic, res => {
              res && res.body && useEmitt().emitter.emit(channel.event, res.body)
            })
          })
        },
        error => {
          disconnect()
          console.error('连接失败: ' + error)
        }
      )
    }

    function disconnect() {
      if (stompClient && stompClient.connected) {
        stompClient.disconnect(
          function () {
            console.info('断开连接')
          },
          function (error) {
            console.info('断开连接失败: ' + error)
          }
        )
      }
      stompClient = null
    }

    function initialize() {
      void connection()
      setInterval(() => {
        if (!isLoginStatus()) {
          disconnect()
          return
        }
        if (!stompClient || !stompClient.connected) {
          void connection()
        }
      }, 5000)
    }
    initialize()
  }
}
