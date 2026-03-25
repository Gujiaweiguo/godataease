import SockJS from 'sockjs-client/dist/sockjs.min.js'
import Stomp from 'stompjs'
import { useCache } from '@/hooks/web/useCache'
import { useEmitt } from '@/hooks/web/useEmitt'
const { wsCache } = useCache()
let stompClient: Stomp.Client
let websocketSupported: boolean | null = null
const env = import.meta.env

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

    function resolvePrefix() {
      let prefix = '/'
      if (window.DataEaseBi?.baseUrl) {
        prefix = window.DataEaseBi.baseUrl
      } else {
        prefix = location.origin + location.pathname
        if (env.MODE === 'dev') {
          prefix = location.origin + '/'
        }
      }
      if (!prefix.endsWith('/')) {
        prefix += '/'
      }
      return prefix
    }

    async function canUseWebsocket(prefix: string, userId: string | number) {
      if (websocketSupported !== null) {
        return websocketSupported
      }
      try {
        const response = await fetch(`${prefix}websocket/info?userId=${userId}`, {
          credentials: 'include'
        })
        if (!response.ok) {
          websocketSupported = false
          return false
        }
        const data = await response.json()
        websocketSupported = data?.websocket !== false
        return websocketSupported
      } catch (error) {
        console.info('websocket capability detection failed:', error)
        websocketSupported = false
        return false
      }
    }

    async function connection() {
      if (!isLoginStatus()) {
        return
      }
      if (stompClient && stompClient.connected) {
        return
      }
      const prefix = resolvePrefix()
      const userId = wsCache.get('app.desktop') ? 1 : wsCache.get('user.uid')
      if (!(await canUseWebsocket(prefix, userId))) {
        disconnect()
        return
      }
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
