import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/Home.vue'
import Commands from '../views/Commands.vue'
import Files from '../views/Files.vue'
import Logs from '../views/logs.vue'
import BashHistory from '../views/BashHistory.vue'
import Agent from '../views/Agent.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/commands',
    name: 'Commands',
    component: Commands
  },
  {
    path: '/files',
    name: 'Files',
    component: Files
  },
  {
    path: '/logs',
    name: 'Logs',
    component: Logs
  },
  {
    path: '/bash-history',
    name: 'BashHistory',
    component: BashHistory
  },
  {
    path: '/agents',
    name: 'Agents',
    component: Agent
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router 