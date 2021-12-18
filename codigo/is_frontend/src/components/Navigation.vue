<template>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <div class="container-fluid">
      <router-link class="navbar-brand" to="/">SIDM</router-link>
      <button
        class="navbar-toggler"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent"
        aria-expanded="false"
        aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav me-auto mb-2 mb-lg-0">
          <li class="nav-item">
            <router-link class="nav-link active" aria-current="page" to="/main">Principal</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/orgLink">Vinculaciones</router-link>
          </li>
          <li class="nav-item dropdown">
          </li>
          <li class="nav-item">
            <router-link class="nav-link disabled" to="#">TODO</router-link>
          </li>
        </ul>
        <ul class="navbar-nav mb-2 mb-lg-0">
          <li v-if="!logged" class="nav-item"><router-link class="btn btn-outline-primary" to="/login">Iniciar sesion</router-link></li>
          <router-link v-if="logged" class="nav-link dropdown-toggle" to="#" id="navbarDropdown" role="button" data-bs-toggle="dropdown" aria-expanded="false">
            {{ userName }}
          </router-link>
          <ul class="dropdown-menu" aria-labelledby="navbarDropdown">
            <li><router-link class="dropdown-item" to="#">Perfil</router-link></li>
            <li><router-link class="dropdown-item" to="#">Cerrar sesion</router-link></li>
          </ul>
        </ul>
      </div>
    </div>
  </nav>
</template>

<script>
export default  {
  name: 'Navigation',
  data() {
    return {
      logged: false,
      userName: '',
    };
  },
}
</script>
