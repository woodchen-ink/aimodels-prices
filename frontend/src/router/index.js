import { createRouter, createWebHistory } from 'vue-router'
import Prices from '../views/Prices.vue'
import Providers from '../views/Providers.vue'
import Login from '../views/Login.vue'
import Home from '../views/Home.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/prices',
      name: 'prices',
      component: Prices
    },
    {
      path: '/providers',
      name: 'providers', 
      component: Providers
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    }
  ]
})

export default router 