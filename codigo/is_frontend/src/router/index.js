import { createRouter, createWebHistory } from 'vue-router';

import Main from '../views/Main';
import ServerLink from '../views/ServerLink';

const routes = [
  {
    path: '/',
    name: 'Main',
    component: Main
  },
  {
    path: '/orgLink',
    name: 'ServerLink',
    component: ServerLink
  }
];

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
});

export default router;
