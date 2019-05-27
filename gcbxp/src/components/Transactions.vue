<template>
  <div class="w3-container w3-center w3-margin">
    <h2 class="w3-margin">{{ txMsg }}</h2>
    <div class="w3-container">
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
          </td>
          <td v-for="vin in tx.Vin" v-bind:key="vin.Txid" class="w3-left-align">
            <p><strong>Txid: </strong>{{vin.Txid}}</p>
            <p><strong>PubKey: </strong>{{vin.PubKey}}</p>
            <p><strong>Signature: </strong>{{vin.Signature}}</p>
          </td>
          <td v-for="vout in tx.Vout" v-bind:key="vout.PubKeyHash" class="w3-left-align">
            <p><strong>Value: </strong>{{vout.Value}}</p>
            <p><strong>PubKeyHash: </strong>{{vout.PubKeyHash}}</p>
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
    return fetch('http://localhost:9050/list_mempool', {
      method: 'get' }).then((res) => res.json())
      .then((data) => {
        console.log(data)
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
