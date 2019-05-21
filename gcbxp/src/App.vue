<template>
  <div id="app">
    <div class="gradient">
      <Nav></Nav>
      <Search></Search>
    </div>
    <div class="w3-section w3-center">
       <router-view></router-view>
    </div>

  </div>
</template>

<script>
import Nav from './components/Nav.vue'
import Search from './components/Search.vue'

export default {
  name: 'app',
  components: {
    'Nav': Nav,
    'Search': Search
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #171219;
}
.gradient {
    background: #193257;
    background: -webkit-linear-gradient(to left, #193257, #0190FF);
    background: -moz-linear-gradient(to left, #193257, #0190FF);
    background: linear-gradient(to left, #193257, #0190FF);
}
</style>
