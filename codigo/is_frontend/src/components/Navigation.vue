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
            <router-link class="nav-link active" aria-current="page" to="/">Principal</router-link>
          </li>
          <li class="nav-item">
            <router-link class="nav-link" to="/orgLink">Vinculaciones</router-link>
          </li>
          <li class="nav-item dropdown">
            <router-link class="nav-link dropdown-toggle" to="#" id="navbarDropdown" role="button" data-bs-toggle="dropdown" aria-expanded="false">
              Usuario
            </router-link>
            <ul class="dropdown-menu" aria-labelledby="navbarDropdown">
              <li><router-link class="dropdown-item" to="#">Perfil</router-link></li>
              <li><router-link class="dropdown-item" to="#">Cerrar sesion</router-link></li>
              <li><hr class="dropdown-divider"></li>
              <li><router-link class="dropdown-item" to="#">Something else here</router-link></li>
            </ul>
          </li>
          <li class="nav-item">
            <router-link class="nav-link disabled" to="#">TODO</router-link>
          </li>
        </ul>
        <button class="btn btn-outline-primary">Login</button>
      </div>
    </div>
  </nav>
</template>

<script>
export default  {
  name: 'Navigation',
}
</script>
