export default {
  server: {
    proxy: {
      '/api/f': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: path => path.replace(/^\/api\/f/, '')
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: path => path
      }
    },
    port: 8080
  }
}
