# Admin Panel Example

This example demonstrates building an admin dashboard with Toutā and Inertia.js.

## Features

- 👥 **User Management** - CRUD operations for users
- 📊 **Dashboard** with real-time metrics
- 🔐 **Role-based Access Control** (RBAC)
- 📈 **Analytics** and charts
- 🔍 **Advanced Search** and filtering
- 📱 **Responsive Design**
- 🌙 **Dark Mode** support
- ⚡ **Real-time Updates** with server-sent events

## Quick Start

```bash
# Generate from ritual
touta ritual init blog

# Choose:
# - Frontend: inertia-vue
# - Enable SSR: yes
# - Authentication: yes
# - Admin Panel: yes
```

## Project Structure

```
admin-panel/
├── handlers/
│   ├── admin/
│   │   ├── dashboard.go      # Dashboard metrics
│   │   ├── users.go          # User management
│   │   ├── settings.go       # System settings
│   │   └── analytics.go      # Analytics data
│   └── middleware/
│       └── admin.go          # Admin auth middleware
├── resources/js/
│   ├── Pages/Admin/
│   │   ├── Dashboard.vue     # Main dashboard
│   │   ├── Users/
│   │   │   ├── Index.vue
│   │   │   ├── Create.vue
│   │   │   └── Edit.vue
│   │   ├── Settings.vue
│   │   └── Analytics.vue
│   └── Components/Admin/
│       ├── Sidebar.vue       # Admin navigation
│       ├── Chart.vue         # Chart component
│       ├── Table.vue         # Data table
│       └── StatsCard.vue     # Metric card
└── models/
    ├── user.go
    └── permission.go
```

## Key Features

### 1. Dashboard with Real-time Metrics

```go
// handlers/admin/dashboard.go
func (h *DashboardHandler) Index(c *cosan.Context) error {
    metrics := h.service.GetMetrics()
    
    return c.Inertia("Admin/Dashboard", map[string]interface{}{
        "metrics": metrics,
        "users": h.service.GetRecentUsers(10),
        "activity": h.service.GetRecentActivity(20),
    })
}
```

```vue
<!-- Pages/Admin/Dashboard.vue -->
<template>
  <AdminLayout>
    <div class="grid grid-cols-4 gap-4">
      <StatsCard
        title="Total Users"
        :value="metrics.totalUsers"
        trend="up"
        :change="metrics.userGrowth"
      />
      <StatsCard
        title="Active Sessions"
        :value="metrics.activeSessions"
        icon="users"
      />
      <StatsCard
        title="Revenue"
        :value="formatCurrency(metrics.revenue)"
        trend="up"
      />
      <StatsCard
        title="Server Load"
        :value="`${metrics.cpuUsage}%`"
        :trend="metrics.cpuUsage > 80 ? 'warning' : 'normal'"
      />
    </div>
    
    <div class="mt-8">
      <Chart
        type="line"
        :data="metrics.userGrowthData"
        title="User Growth (30 days)"
      />
    </div>
  </AdminLayout>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import AdminLayout from '@/Components/Admin/Layout.vue'
import StatsCard from '@/Components/Admin/StatsCard.vue'
import Chart from '@/Components/Admin/Chart.vue'

const props = defineProps({
  metrics: Object,
  users: Array,
  activity: Array
})

// Real-time updates via SSE
let eventSource

onMounted(() => {
  eventSource = new EventSource('/admin/metrics/stream')
  eventSource.onmessage = (event) => {
    const newMetrics = JSON.parse(event.data)
    Object.assign(props.metrics, newMetrics)
  }
})

onUnmounted(() => {
  eventSource?.close()
})
</script>
```

### 2. User Management with Advanced Table

```vue
<!-- Pages/Admin/Users/Index.vue -->
<template>
  <AdminLayout>
    <div class="flex justify-between mb-4">
      <h1 class="text-2xl font-bold">Users</h1>
      <Link href="/admin/users/create" class="btn-primary">
        Add User
      </Link>
    </div>
    
    <div class="mb-4 flex gap-4">
      <input
        v-model="search"
        type="search"
        placeholder="Search users..."
        class="input"
      />
      <select v-model="roleFilter" class="select">
        <option value="">All Roles</option>
        <option value="admin">Admin</option>
        <option value="editor">Editor</option>
        <option value="user">User</option>
      </select>
    </div>
    
    <Table
      :columns="columns"
      :data="filteredUsers"
      :sortable="true"
      @row-click="editUser"
    >
      <template #actions="{ row }">
        <button @click="deleteUser(row.id)" class="btn-danger-sm">
          Delete
        </button>
      </template>
    </Table>
    
    <Pagination
      :current="users.current_page"
      :total="users.total_pages"
      @change="loadPage"
    />
  </AdminLayout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Link, router } from '@toutaio/inertia-vue'
import AdminLayout from '@/Components/Admin/Layout.vue'
import Table from '@/Components/Admin/Table.vue'
import Pagination from '@/Components/Pagination.vue'

const props = defineProps({
  users: Object
})

const search = ref('')
const roleFilter = ref('')

const columns = [
  { key: 'id', label: 'ID', sortable: true },
  { key: 'name', label: 'Name', sortable: true },
  { key: 'email', label: 'Email', sortable: true },
  { key: 'role', label: 'Role', sortable: true },
  { key: 'created_at', label: 'Joined', sortable: true },
  { key: 'actions', label: 'Actions' }
]

const filteredUsers = computed(() => {
  let filtered = props.users.data
  
  if (search.value) {
    filtered = filtered.filter(u =>
      u.name.toLowerCase().includes(search.value.toLowerCase()) ||
      u.email.toLowerCase().includes(search.value.toLowerCase())
    )
  }
  
  if (roleFilter.value) {
    filtered = filtered.filter(u => u.role === roleFilter.value)
  }
  
  return filtered
})

const editUser = (user) => {
  router.visit(`/admin/users/${user.id}/edit`)
}

const deleteUser = (id) => {
  if (confirm('Are you sure?')) {
    router.delete(`/admin/users/${id}`, {
      onSuccess: () => {
        alert('User deleted successfully')
      }
    })
  }
}

const loadPage = (page) => {
  router.visit(`/admin/users?page=${page}`)
}
</script>
```

### 3. Role-Based Access Control

```go
// handlers/middleware/admin.go
func RequireAdmin(c *cosan.Context) error {
    user := c.Get("user").(*models.User)
    
    if !user.HasRole("admin") {
        return c.Redirect("/", 403)
    }
    
    return c.Next()
}

func RequirePermission(permission string) cosan.HandlerFunc {
    return func(c *cosan.Context) error {
        user := c.Get("user").(*models.User)
        
        if !user.HasPermission(permission) {
            return c.JSON(403, map[string]string{
                "error": "Forbidden",
            })
        }
        
        return c.Next()
    }
}

// main.go - Apply middleware
admin := router.Group("/admin")
admin.Use(middleware.RequireAuth, middleware.RequireAdmin)

admin.GET("/dashboard", dashboardHandler.Index)
admin.GET("/users", usersHandler.Index)
admin.POST("/users", usersHandler.Create, 
    middleware.RequirePermission("users.create"))
```

### 4. Analytics Dashboard

```vue
<!-- Pages/Admin/Analytics.vue -->
<template>
  <AdminLayout>
    <h1 class="text-2xl font-bold mb-6">Analytics</h1>
    
    <div class="grid grid-cols-2 gap-6">
      <div class="card">
        <h2 class="text-lg font-semibold mb-4">Page Views</h2>
        <Chart
          type="bar"
          :data="analytics.pageViews"
          :options="chartOptions"
        />
      </div>
      
      <div class="card">
        <h2 class="text-lg font-semibold mb-4">Traffic Sources</h2>
        <Chart
          type="pie"
          :data="analytics.trafficSources"
        />
      </div>
      
      <div class="card">
        <h2 class="text-lg font-semibold mb-4">User Retention</h2>
        <Chart
          type="line"
          :data="analytics.retention"
        />
      </div>
      
      <div class="card">
        <h2 class="text-lg font-semibold mb-4">Top Pages</h2>
        <ul class="space-y-2">
          <li v-for="page in analytics.topPages" :key="page.url">
            <div class="flex justify-between">
              <span>{{ page.url }}</span>
              <span class="font-semibold">{{ page.views }}</span>
            </div>
            <div class="w-full bg-gray-200 h-2 rounded">
              <div
                class="bg-blue-600 h-2 rounded"
                :style="{ width: `${page.percentage}%` }"
              ></div>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </AdminLayout>
</template>

<script setup>
import { ref } from 'vue'
import AdminLayout from '@/Components/Admin/Layout.vue'
import Chart from '@/Components/Admin/Chart.vue'

defineProps({
  analytics: Object
})

const chartOptions = {
  responsive: true,
  plugins: {
    legend: {
      display: true
    }
  }
}
</script>
```

## Development

### Start development server

```bash
# Terminal 1: Frontend dev server
npm run dev

# Terminal 2: Go server with hot reload
air
```

### Run tests

```bash
# Go tests
go test ./...

# Vue tests
npm test
```

## Customization

### Add new admin section

1. Create handler:

```go
// handlers/admin/reports.go
func (h *ReportHandler) Index(c *cosan.Context) error {
    reports := h.service.GetReports()
    return c.Inertia("Admin/Reports/Index", map[string]interface{}{
        "reports": reports,
    })
}
```

2. Create Vue page:

```vue
<!-- Pages/Admin/Reports/Index.vue -->
<template>
  <AdminLayout>
    <h1>Reports</h1>
    <!-- Your content -->
  </AdminLayout>
</template>
```

3. Add route:

```go
admin.GET("/reports", reportHandler.Index)
```

4. Add to sidebar navigation in `Components/Admin/Sidebar.vue`

## Deployment

See main documentation for deployment instructions. Admin panel is deployed as part of the main application.

## Security

- ✅ CSRF protection enabled
- ✅ XSS protection with Vue
- ✅ SQL injection prevention with prepared statements
- ✅ Session security with httpOnly cookies
- ✅ Rate limiting on sensitive endpoints
- ✅ Audit logging for admin actions

## License

MIT
