const state = {
  files: {
    '9': {
      'serverName': 'informe1.pdf',
      'patient': 'david brent',
      'updatedAt': '16/12/2018',
      'fetchToken': 'asdqwe',
    }
  },
};

const getters = {
  fileInfo: (state) => (id) => (id in state.files ? state.files[id] : null),
};

const actions = {};

const mutations = {};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations,
};
