<template>
  <li>
    <div :class="[isFolder ? 'folder' : 'file']" @click="onClick(root)">
      {{root.text}}
      <span v-if="isFolder">[{{ isOpen ? '-' : '+' }}]</span>
    </div>
    <ul v-show="isOpen" v-if="isFolder">
      <Tree
        class="item"
        v-for="(child, index) in root.children" :key="index"
        :root="child"
      />
    </ul>
  </li>
</template>

<script>

import { mapActions } from 'vuex';

export default {
  name: 'Tree',
  data() {
    return {
      isOpen: false
    }
  },
  props: {
    root: Object,
  },
  computed: {
    isFolder() {
      return this.root.children && this.root.children.length;
    }
  },
  methods: {
    ...mapActions('mappings', ['updateSelected']),
    toggle() {
      if (this.isFolder) {
        this.isOpen = !this.isOpen;
      }
    },
    onClick(node) {
      if (this.isFolder) {
        this.toggle();
      } else {
        this.updateSelected(node);
      }
    },
  }
}
</script>

<style scoped>
  .item {
    cursor: pointer
  }
  .folder {
    color: green
  }
  .file {
    color: blue
  }
  .ul {
    padding-left: 1em;
    line-height: 1.5em;
    list-style-type: dot;
  }
</style>
