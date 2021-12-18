import { createRouter, createWebHistory } from 'vue-router';

import Main from '@/views/Main';
import ServerLink from '@/views/ServerLink';
import Login from '@/views/Login';
import About from '@/views/About';

const routes = [
  {
    path: '/',
    name: 'About',
    component: About
  },
  {
    path: '/main',
    name: 'Main',
    component: Main
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
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
