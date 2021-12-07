const emptyFileInfo = {
  fsData : {
    serverName: 'N/A',
    patient: 'N/A',
    updatedAt: 'N/A',
    fetchToken: 'N/A',
  },
  isData: {
    path: 'N/A',
  }
};

const addRoot = (mappings) => ({
  'id': '-1',
  'text': 'root',
  'path': '',
  'children': mappings
})

export default {
  namespaced: true,
  state: {
    mappings: [],
    selected: emptyFileInfo,
  },
  getters: {
    allMappings: (state) => state.mappings,
    selectedMapping: (state) => state.selected,
  },
  actions: {
    async fetchMappings({ commit }) {
      const resp = await fetch('http://localhost:9876/main/mappings');
      const mappings = await resp.json()
      commit('setMappings', mappings);
    },

    updateSelected({ commit, rootGetters }, selected) {
      const fi = rootGetters['files/fileInfo'](selected.id);
      commit('setSelected', null == fi 
        ? emptyFileInfo
        : { fsData: fi, isData: selected, });
    }
  },
  mutations: {
    setMappings: (state, mappings) => (state.mappings = addRoot(mappings)),
    setSelected: (state, selected) => (state.selected = selected),
  },
};
