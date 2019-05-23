<template>
  <div class="w3-container w3-center">
    <h2 class="w3-margin">{{ blocksMsg }}</h2>
    <div class="w3-content">
      <table class="w3-table w3-bordered w3-centered w3-striped">
        <tr>
          <th>Height</th>
          <th>Timestamp</th>
          <th>Transactions</th>
        </tr>
        <tr>

        </tr>
      </table>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Blocks',
  props: {
    msg: String
  },
  data () {
    return {
      blocksMsg: 'Latest Blocks'
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
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
