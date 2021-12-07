<template>
  <div class="container">
    <div class="row">
      <Navigation />
    </div>
    <div class="row">
      <router-view />
    </div>
  </div>
</template>

<script>

import Navigation from './components/Navigation'

export default {
  name: 'App',
  components: {
    Navigation
  },
}
</script>
