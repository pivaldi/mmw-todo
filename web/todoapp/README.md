# Todo Manager - Angular Frontend

A modern, responsive Angular web application for managing todos, built to work with the DDD/Hexagonal Architecture backend.

## Features

- **Complete CRUD Operations** - Create, read, update, and delete todos
- **Status Management** - Track todos as Pending, In Progress, Completed, or Cancelled
- **Priority Levels** - Organize by Low, Medium, High, or Urgent priority
- **Due Date Tracking** - Set and monitor due dates with visual indicators
- **Filtering** - Filter todos by status and priority
- **Responsive Design** - Works seamlessly on desktop, tablet, and mobile
- **Real-time Updates** - Immediate feedback on all operations

## Tech Stack

- **Angular 17** - Modern web framework
- **TypeScript** - Type-safe development
- **RxJS** - Reactive programming
- **Angular Router** - Client-side routing
- **HttpClient** - API communication

## Prerequisites

- Node.js 18+ and npm
- Running Todo API backend (see main project README)

## Getting Started

### 1. Install Dependencies

```bash
cd web/todoapp
npm install
```

### 2. Configure API Endpoint

The application is configured to proxy API requests to `http://localhost:8090` by default. If your backend runs on a different port, update `proxy.conf.json`:

```json
{
  "/api": {
    "target": "http://localhost:YOUR_PORT",
    "secure": false,
    "changeOrigin": true
  }
}
```

### 3. Start the Development Server

```bash
npm start
```

The application will be available at `http://localhost:4200`

### 4. Build for Production

```bash
npm run build
```

Production files will be output to `dist/todoapp/`

## Project Structure

```
src/
├── app/
│   ├── components/
│   │   ├── todo-list/       # List view with filtering
│   │   ├── todo-detail/     # Detailed todo view
│   │   └── todo-form/       # Create/edit form
│   ├── models/
│   │   └── todo.model.ts    # TypeScript interfaces
│   ├── services/
│   │   └── todo.service.ts  # API communication
│   ├── app.component.*      # Root component
│   ├── app.module.ts        # App module
│   └── app-routing.module.ts # Routing config
├── environments/
│   └── environment.ts       # Environment config
├── assets/                  # Static assets
├── index.html              # HTML entry point
├── main.ts                 # Bootstrap file
└── styles.css              # Global styles
```

## Components

### TodoList Component

- Displays all todos in a responsive grid
- Filters by status and priority
- Quick actions: complete, delete
- Click to view details

### TodoDetail Component

- Full todo information
- Status and priority badges
- Due date warnings (overdue, due soon)
- Actions: edit, complete, reopen, delete

### TodoForm Component

- Create new todos
- Edit existing todos
- Form validation
- Due date picker

## API Integration

The application communicates with the backend via HTTP:

- `GET /api/todo.v1.TodoService/ListTodos` - List todos
- `GET /api/todo.v1.TodoService/GetTodo` - Get single todo
- `POST /api/todo.v1.TodoService/CreateTodo` - Create todo
- `PUT /api/todo.v1.TodoService/UpdateTodo` - Update todo
- `POST /api/todo.v1.TodoService/CompleteTodo` - Mark complete
- `POST /api/todo.v1.TodoService/ReopenTodo` - Reopen todo
- `DELETE /api/todo.v1.TodoService/DeleteTodo` - Delete todo

## Development

### Code Formatting

```bash
npm run lint
```

### Running Tests

```bash
npm test
```

### Building

```bash
# Development build
npm run build

# Production build with optimizations
npm run build -- --configuration production
```

## Deployment

### Static Hosting

Build the application and serve the `dist/todoapp` directory:

```bash
npm run build -- --configuration production
```

Deploy the contents of `dist/todoapp/` to any static hosting service:
- Nginx
- Apache
- AWS S3 + CloudFront
- Netlify
- Vercel

### Nginx Configuration Example

```nginx
server {
    listen 80;
    server_name todoapp.example.com;
    root /var/www/todoapp;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Features Detail

### Status Indicators

- **Pending** - Gray badge
- **In Progress** - Blue badge
- **Completed** - Green badge
- **Cancelled** - Light gray badge

### Priority Indicators

- **Low** - Light blue badge
- **Medium** - Orange badge
- **High** - Dark orange badge
- **Urgent** - Red badge

### Due Date Warnings

- **Overdue** - Red text with warning
- **Due Soon** (< 24 hours) - Orange text with warning
- **Future** - Normal text

## Troubleshooting

### API Connection Issues

1. Ensure backend is running on port 8090
2. Check `proxy.conf.json` configuration
3. Verify CORS settings on backend

### Build Errors

1. Clear node_modules: `rm -rf node_modules && npm install`
2. Clear cache: `npm cache clean --force`
3. Check Node.js version: `node --version` (should be 18+)

## License

See main project LICENSE file.
