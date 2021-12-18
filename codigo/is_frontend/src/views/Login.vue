<template>
  <div>
    <h1>IsAuthorized: {{ isLoggedIn }}</h1>
    <!-- <h2 v-if="user">signed user: {{user}}</h2> -->
    <button @click="handleClickSignIn" :disabled="isLoggedIn">sign in</button>
    <button @click="handleClickSignOut" :disabled="!isLoggedIn">sign out</button>
  </div>

</template>

<script>

import { inject } from 'vue';
import { mapGetters, mapActions } from 'vuex';

export default {
  name: 'Login',
  computed: {
    ...mapGetters('login', ['isLoggedIn']),
  },
  methods: {
    ...mapActions('login', ['login', 'logout']),
    async handleClickSignIn(){
      try {
        const googleUser = await this.$gAuth.signIn();
        if (!googleUser) {
          return null;
        }
        console.log("googleUser", googleUser);
        this.login(googleUser);
      } catch (error) {
        //on fail do something
        console.error(error);
        return null;
      }
    },
    async handleClickSignOut() {
      try {
        await this.$gAuth.signOut();
        console.log("isAuthorized", this.googleOAuth.isAuthorized);
        this.logout();
      } catch (error) {
        console.error(error);
      }
    },
  },
  setup() {
    const googleOAuth = inject("Vue3GoogleOauth");
    return { googleOAuth };
  },
}
</script>
