<template>
  <div class="wrapper">
   
    <side-bar>
      <template slot="links">
        <sidebar-item
          :link="{
            name: 'Dashboard',
            path: '/dashboard',
            icon: 'ni ni-tv-2 text-primary',
          }"
        >
        </sidebar-item>

        <sidebar-item
            :link="{
              name: 'Agents',
              path: '/agents',
              icon: 'ni ni-vector'
              }"
            >
        </sidebar-item>

        <sidebar-item
                :link="{
                  name: 'Logs',
                  path: '/logs',
                  icon: 'ni ni-bullet-list-67 text-red'
                }">
        </sidebar-item>


        <sidebar-item
              :link="{
                name: 'Reports',
                path: '/maps',
                icon: 'ni ni-sound-wave text-orange'
              }">
        </sidebar-item>

        
        <sidebar-item
                :link="{
                  name: 'Commands',
                  path: '/commands',
                  icon: 'ni ni-bullet-list-67 text-red'
                }">
        </sidebar-item>

        <sidebar-item
                :link="{
                  name: 'Tables',
                  path: '/tables',
                  icon: 'ni ni-bullet-list-67 text-red'
                }">
        </sidebar-item>

        <sidebar-item
                  :link="{
                    name: 'Login',
                    path: '/login',
                    icon: 'ni ni-key-25 text-info'
                  }">
        </sidebar-item>

      </template>
    </side-bar>
    <div class="main-content">
     

      <div @click="$sidebar.displaySidebar(false)">
        <fade-transition :duration="200" origin="center top" mode="out-in">
          <!-- your content here -->
          <router-view></router-view>
        </fade-transition>
      </div>
      
    </div>
  </div>
  
</template>


<script>
  /* eslint-disable no-new */
  import PerfectScrollbar from 'perfect-scrollbar';
  import 'perfect-scrollbar/css/perfect-scrollbar.css';

  function hasElement(className) {
    return document.getElementsByClassName(className).length > 0;
  }

  function initScrollbar(className) {
    if (hasElement(className)) {
      new PerfectScrollbar(`.${className}`);
    } else {
      // try to init it later in case this component is loaded async
      setTimeout(() => {
        initScrollbar(className);
      }, 100);
    }
  }

  import DashboardNavbar from './DashboardNavbar.vue';
  import ContentFooter from './ContentFooter.vue';
  import DashboardContent from './Content.vue';
  import { FadeTransition } from 'vue2-transitions';

  export default {
    components: {
      DashboardNavbar,
      ContentFooter,
      DashboardContent,
      FadeTransition
    },
    methods: {
      initScrollbar() {
        let isWindows = navigator.platform.startsWith('Win');
        if (isWindows) {
          initScrollbar('sidenav');
        }
      }
    },
    mounted() {
      this.initScrollbar()
    }
  };
</script>
<style lang="scss">
  .side-bar {
    background-color: #1a1f36 !important; /* Dark bluish shade */
    width: 200px !important; /* Reduce width */
    min-width: 200px !important;
  }

  .main-content {
    margin-left: 200px !important; /* Adjust main content spacing */
  }

  .side-bar .sidebar-item {
    color: #ffffff; /* White text for contrast */
    padding: 10px 15px; /* Adjust padding */
  }

  .side-bar .sidebar-item:hover {
    background-color: #2a2f48; /* Slightly lighter shade for hover */
  }

  .side-bar .sidebar-item .ni {
    color: #ffffff !important; /* Ensuring icons remain visible */
  }
</style>
