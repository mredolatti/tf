import { createStore } from 'vuex'

import login from './modules/login';
import mappings from './modules/mappings';
import files from './modules/files';

export default createStore({
  modules: {
    login,
    files,
    mappings,
  }
});
