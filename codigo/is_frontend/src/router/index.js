import { createRouter, createWebHistory } from 'vue-router';

import ServerLink from '../views/ServerLink';

const routes = [
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
