<template>
    <div class="row">
      <div class="col">
        <h4>Vista de archivos</h4>
        <ul>
          <Tree :root=allMappings class="item" />
        </ul>
      </div>
      <div class="col">
        <h4>Informacion de archivo seleccionado</h4>
        <FileInfo />
      </div>
    </div>
</template>

<script>

import Tree from '../components/Tree';
import FileInfo from '../components/FileInfo';
import { mapGetters, mapActions } from 'vuex';

export default {
  name: 'Main',
  components: {
    Tree,
    FileInfo
  },
  data() {
    return {};
  },
  methods: {
    ...mapActions('mappings', ['fetchMappings']),
  },
  computed: {
    ...mapGetters('mappings', ['allMappings']),
  },
  async created() {
    this.fetchMappings();
  }
}
</script>
