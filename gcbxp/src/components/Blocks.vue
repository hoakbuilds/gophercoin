<template>
  <div class="w3-container w3-center">
    <h2 class="w3-margin">{{ blocksMsg }}</h2>
    <div class="w3-content">
      <table class="w3-table w3-bordered w3-centered w3-striped">
        <tr>
          <th>Timestamp</th>
          <th>Height</th>
          <th>Hash</th>
          <th>Transactions</th>
        </tr>
        <tr v-for="block in blocks" v-bind:key="block.height">
          <td>{{block.Timestamp}}</td>
          <td>{{block.Height}}</td>
          <td>{{block.Hash}}</td>
          <td>{{block.TXcount}}</td>
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
      blocksMsg: 'Latest Blocks',
      blocks: null
    }
  },
  mounted () {
    return fetch('http://localhost:9050/list_blocks', {
      method: 'get' }).then((res) => res.json())
      .then((data) => {
        var i
        console.log(data.Blocks)
        for (i = 0; i < data.Blocks.length; i++) {
          data.Blocks[i].TXcount = data.Blocks[i].Transactions.length
        }
        this.blocks = data.Blocks
      }).catch((err) => console.log(err))
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
