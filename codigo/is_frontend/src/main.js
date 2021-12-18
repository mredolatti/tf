import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import GAuth from 'vue3-google-oauth2'

import 'bootstrap/dist/css/bootstrap.min.css';

const gAuthOptions = {
    clientId: '814369198327-2mvjfu9nt1h3prthspe71papgjnborr0.apps.googleusercontent.com',
    scope: 'email profile openid',
}

createApp(App)
    .use(store)
    .use(router)
    .use(GAuth, gAuthOptions)
    .mount('#app');
