<template>
  <table class="table">
    <tr>
      <th>Ruta completa</th>
      <td>{{ selectedMapping.isData.path }}</td>
    </tr>
    <tr>
      <th>Archivo remoto</th>
      <td>{{ selectedMapping.fsData.serverName }}</td>
    </tr>
    <tr>
      <th>Paciente</th>
      <td>{{ selectedMapping.fsData.patient }}</td>
    </tr>
    <tr>
      <th>Ultima actualizacion</th>
      <td>{{ selectedMapping.fsData.updatedAt }}</td>
    </tr>
    <tr>
      <th>Operaciones</th>
      <td><button type="button" class="btn btn-outline-primary">Descargar</button></td>
      <td><button type="button" class="btn btn-outline-primary">Mover</button></td>
    </tr>
  </table>
</template>

<script>

import { mapGetters } from 'vuex';

export default {
  name: 'FileInfo',
  computed: {
    ...mapGetters('mappings', ['selectedMapping']),
  }
}
</script>

<style scoped>

</style>
