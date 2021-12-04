<template>
  <table class="table">
    <tr>
      <th>Archivo remoto</th>
      <td>{{ file.serverName }}</td>
    </tr>
    <tr>
      <th>Paciente</th>
      <td>{{ file.patient }}</td>
    </tr>
    <tr>
      <th>Ultima actualizacion</th>
      <td>{{ file.updatedAt }}</td>
    </tr>
    <tr>
      <th>Descarga</th>
      <td><a href="#">Descargar archivo</a></td>
    </tr>
  </table>
  <!--
  <ul class="list-group list-group-flush">
    <li class="list-group-item"><p><b>Archivo remoto</b> {{ file.serverName }}</p></li>
    <li class="list-group-item"><p><b>Paciente</b> {{ file.patient }} </p></li>
    <li class="list-group-item"><p><b>Ultima actualizacion</b> {{ file.updatedAt }}</p></li>
    <li class="list-group-item"><p><b>Descarga</b> <a href="#">Descargar archivo</a></p></li>
  </ul>
  -->
</template>

<script>
export default {
  name: 'FileInfo',
  props: {
    file: Object
  },
}
</script>
