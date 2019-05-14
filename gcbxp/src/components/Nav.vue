<template>
  <div class="w3-center w3-container gradient">
    <nav class="w3-section w3-bar">
      <router-link  class="w3-btn w3-wide" v-for="routes in links"
      v-bind:key="routes.id"
      :to="`${routes.page}`">
      {{routes.text}}
      </router-link>
    </nav>
  </div>
</template>
<script>
export default {
  name: 'Nav',
  data () {
    return {
      links: [
        {
          id: 0,
          text: 'Index',
          page: '/'
        },
        {
          id: 1,
          text: 'Blocks',
          page: '/blocks'
        },
        {
          id: 2,
          text: 'Transactions',
          page: '/transactions'
        }
      ]
    }
  }
}

</script>

<style>
  .spacing {
    margin-right: 10px;
  }
  .router-link-active {
     font-weight: 500;
  }

</style>
