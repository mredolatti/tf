<template>
  <div v-for="file in files" :key="file.id">
    <h3>{{ file.text }}</h3>
  </div>
</template>

<script>
export default {
  name: 'FileTree',
  props: {
    files: Array
  }
}
</script>
