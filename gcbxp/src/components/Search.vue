<template>
  <div class="w3-section w3-margin-bottom">
    <div class="w3-section w3-center w3-padding-16">
      <h1><img class="w3-image" src="../assets/animated_color.gif" style="max-height: 66px;">Block Explorer</h1>
    </div>
    <div class="w3-center" style="padding-bottom:75px;">
        <!-- The form -->
        <form class="w3-container" action="">
        <input class="w3-input w3-half" style="float:none; width:40%; display:inline;" type="text" placeholder="Search for block height, hash, transaction, or address" name="search">
        </form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Search',
  props: {
    msg: String
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
input[type=text], textarea {
  -webkit-transition: all 0.30s ease-in-out;
  -moz-transition: all 0.30s ease-in-out;
  -ms-transition: all 0.30s ease-in-out;
  -o-transition: all 0.30s ease-in-out;
  outline: none;
  padding: 3px 0px 3px 3px;
  margin: 5px 1px 3px 0px;
  border: 1px solid #DDDDDD;
}

input[type=text]:focus, textarea:focus {
  box-shadow: 3 3 9px rgb(238, 81, 81);
  padding: 3px 0px 3px 3px;
  margin: 5px 1px 3px 0px;
  border: 1px solid rgb(238, 131, 81);
}
</style>
