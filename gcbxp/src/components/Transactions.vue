<template>
  <div class="w3-container w3-center w3-margin">
    <h2 class="w3-margin">{{ txMsg }}</h2>
    <div class="w3-content">
      <table class="w3-table w3-bordered w3-centered w3-striped">
      <tr>
        <th>Transaction ID</th>
        <th>Value</th>
      </tr>
      <tr>
      </tr>
      </table>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Transactions',
  data () {
    return {
      txMsg: 'Transaction Mempool'
    }
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
