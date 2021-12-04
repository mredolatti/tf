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
        @display-file-info="propagateDisplayInfo"
      />
    </ul>
  </li>
</template>

<script>
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
    toggle() {
      if (this.isFolder) {
        this.isOpen = !this.isOpen;
      }
    },
    onClick(node) {
      if (this.isFolder) {
        this.toggle();
      } else {
        this.$emit('display-file-info', node);
      }
    },
    propagateDisplayInfo(data) {
      this.$emit("display-file-info", data);
    }
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
