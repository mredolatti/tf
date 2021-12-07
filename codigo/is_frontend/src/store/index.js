import { createStore } from 'vuex'

import mappings from './modules/mappings';
import files from './modules/files';

export default createStore({
  modules: {
    files,
    mappings,
  }
});
