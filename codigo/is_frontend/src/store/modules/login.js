import { inject } from 'vue';

export default {
  namespaced: true,
  state: {
    auth: inject("Vue3GoogleOauth"),
    isLoggedIn: false,
    userInfo : {},
  },
  getters: {
    isLoggedIn: (state) => (state.isLoggedIn),
    getUserInfo: (state) => (state.userInfo),
  },
  actions: {
    login({ commit }, userInfo) {
      console.log("AAAAAA", this.auth);
      commit('updateLoginStatus', true);
      console.log(userInfo);
    },
    logout({ commit }) {
      commit('updateLoginStatus', false);
    }
  },
  mutations: {
    updateUserProfile: (state, profile) => (state.userInfo = profile),
    updateLoginStatus: (state, status) => (state.isLoggedIn = status),
  },
};
