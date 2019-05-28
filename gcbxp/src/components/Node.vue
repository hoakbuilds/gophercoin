<template>
  <div class="w3-content w3-center w3-padding-24">
    <div class="w3-card-4 w3-container w3-center">
      <div class="w3-margin w3-padding w3-center">
        <div class="w3-left w3-half w3-padding-16">
          <p><strong>Protocol Port:</strong></p>
          <p><strong>Connected:</strong></p>
          <p><strong>Mining:</strong></p>
          <p><strong>Synced:</strong></p>
        </div>
        <div class="w3-right w3-half w3-padding-16">
          <p>{{info.Protocol}}</p>
          <p>{{info.Peers}}</p>
          <p>{{info.Mining}}</p>
          <p>{{info.Synced}}</p>
        </div>
      </div>
    </div>
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
    return fetch('http://localhost:9050//node_info', {
      method: 'get' }).then((res) => res.json())
      .then((data) => {
        console.log(data)
        this.info = data
      }).catch((err) => console.log(err))
  }
}
</script>
