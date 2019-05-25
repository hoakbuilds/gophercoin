<template>
  <div class="w3-container w3-center w3-margin">
    <h2 class="w3-margin">{{ txMsg }}</h2>
    <div class="w3-content">
      <table class="w3-table w3-bordered w3-centered w3-striped">
      <tr>
        <th>Transaction ID</th>
        <th>Vin</th>
        <th>Vout</th>
      </tr>
      <tr v-for="tx in txs"
          v-bind:key="tx.ID">
          <td>
            {{tx.ID}}
            {{tx.Vin}}
          </td>
          <td v-for="vin in tx.Vout" v-bind:key="vin.PubKeyHash">
            {{vin.Txid}}
            {{vin.PubKey}}
          </td>
          <td v-for="vout in tx.Vout" v-bind:key="vout.PubKeyHash">
            {{vout.Value}}
            {{vout.PubKeyHash}}
          </td>
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
      txMsg: 'Transaction Mempool',
      txs: null,
      txCount: null
    }
  },
  mounted () {
    return fetch('/api/list_mempool', {
      method: 'get' }).then((res) => res.json())
      .then((data) => {
        console.log(data.Transactions)
        this.txCount = data.Transactions.length
        this.txs = data.Transactions
      }).catch((err) => console.log(err))
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
