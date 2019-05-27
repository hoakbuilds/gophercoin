<template>
  <div class="w3-container w3-center">
    <h2 class="w3-margin">{{ indexMsg }}</h2>
    <Node></Node>
  </div>
</template>

<script>

import Node from './Node'

export default {
  name: 'Index',
  components: {
    'Node': Node
  },
  data () {
    return {
      indexMsg: 'Node Information'
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
