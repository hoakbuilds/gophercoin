<template>
  <div class="w3-container w3-center">
      {{info}}
  </div>
</template>

<script>
export default {
  name: 'Node',
  data () {
    return {
      info: null
    }
  },
  mounted () {
    return fetch('/api/node_info', {
      method: 'get' }).then((res) => res.json())
      .then((data) => {
        console.log(data)
        this.info = data
      }).catch((err) => console.log(err))
  }
}
</script>
