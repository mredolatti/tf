<template>
  <div class="container">
    <div class="row">
      <Navigation />
    </div>
    <router-view />
    <div class="row">
      <div class="col">
        <h4>Vista de archivos</h4>
        <ul>
          <Tree :root="fileRoot" @display-file-info="displayFileInfo" class="item" />
        </ul>
      </div>
      <div class="col">
        <h4>Informacion de archivo seleccionado</h4>
        <FileInfo :file="selected" />
      </div>
    </div>
  </div>
</template>

<script>
import Navigation from './components/Navigation'
import Tree from './components/Tree'
import FileInfo from './components/FileInfo'

export default {
  name: 'App',
  components: {
    // Header,
    Tree,
    FileInfo,
    Navigation
  },
  data() {
    return {
      fileRoot: {'id': '125', 'text': 'root', 'type': 'folder', 'children': []},
      default: {
        'serverName': 'N/A',
        'patient': 'N/A',
        'updatedAt': 'N/A',
        'fetchToken': 'N/A',
      },
      selected: {},
      fileData: {
        '9': {
          'serverName': 'informe1.pdf',
          'patient': 'david brent',
          'updatedAt': '16/12/2018',
          'fetchToken': 'asdqwe',
        }
      }
    }
  },
  methods: {
    displayFileInfo(file) {
      if (file.id in this.fileData) {
        this.selected = this.fileData[file.id];
      } else {
        this.selected = this.default;
      }
    },
    async fetchMappings() {
      const resp = await fetch('http://localhost:9876/main/mappings');
      const data = await resp.json();
      return data;
    }
  },
  async created() {
    const data = await this.fetchMappings();
    this.fileRoot = {
      'id': '-1',
      'text': 'root',
      'type': 'folder',
      'children': data
    };
  }
}
</script>
