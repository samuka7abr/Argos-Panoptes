import { NavLink } from "./NavLink";
import { Activity, Server, Database, Globe, Mail, Shield, LayoutDashboard, Bell } from "lucide-react";

interface LayoutProps {
  children: React.ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  const navItems = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard },
    { to: "/webserver", label: "Web Server", icon: Server },
    { to: "/database", label: "Database", icon: Database },
    { to: "/dns", label: "DNS", icon: Globe },
    { to: "/smtp", label: "SMTP", icon: Mail },
    { to: "/alerts", label: "Alertas", icon: Bell },
    { to: "/security", label: "Segurança", icon: Shield },
  ];

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b border-border bg-card/80 backdrop-blur-sm">
        <div className="container flex h-16 items-center justify-between">
          <div className="flex items-center gap-2">
            <Activity className="h-6 w-6 text-primary" />
            <h1 className="text-xl font-bold text-gradient">DevOps Monitor</h1>
          </div>
          <div className="text-sm text-muted-foreground">
            Backend: <span className="text-primary font-medium">Golang</span>
          </div>
        </div>
      </header>

      {/* Navigation */}
      <nav className="border-b border-border bg-card/50">
        <div className="container">
          <div className="flex gap-1 overflow-x-auto py-2">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className="flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                activeClassName="bg-primary text-primary-foreground hover:bg-primary/90"
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="container py-6">{children}</main>
    </div>
  );
};

export default Layout;
