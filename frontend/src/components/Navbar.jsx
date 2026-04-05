import { NavLink } from "react-router-dom";
import { LogOut, PieChart, Users, WalletCards } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import RoleGuard from "@/components/RoleGuard";

const navClass = ({ isActive }) =>
  [
    "rounded-md px-3 py-2 text-sm font-medium transition-colors",
    isActive ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:text-foreground hover:bg-muted/30",
  ].join(" ");

export default function Navbar() {
  const { user, logout } = useAuth();

  return (
    <>
      <header className="sticky top-0 z-20 border-b border-border/80 bg-background/90 backdrop-blur md:hidden">
        <div className="mx-auto flex w-full items-center justify-between px-4 py-3 sm:px-6">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-ash-300">FinDash</p>
            <h1 className="text-base font-semibold text-foreground">Finance Control Panel</h1>
          </div>
          <Button variant="secondary" size="sm" onClick={logout}>
            <LogOut className="h-4 w-4" />
            Logout
          </Button>
        </div>
        <nav className="mx-auto flex w-full items-center gap-2 overflow-x-auto px-4 pb-3 sm:px-6">
          <NavLink to="/dashboard" className={navClass}>
            Dashboard
          </NavLink>
          <NavLink to="/entries" className={navClass}>
            Entries
          </NavLink>
          <RoleGuard allowedRoles={["admin", "manager"]}>
            <NavLink to="/analytics" className={navClass}>
              Analytics
            </NavLink>
          </RoleGuard>
          <RoleGuard allowedRoles={["admin", "manager"]}>
            <NavLink to="/users" className={navClass}>
              Users
            </NavLink>
          </RoleGuard>
        </nav>
      </header>

      <aside className="hidden h-screen border-r border-border/80 bg-ash-900/80 md:sticky md:top-0 md:flex md:flex-col">
        <div className="border-b border-border/80 px-5 py-5">
          <p className="text-xs uppercase tracking-[0.25em] text-ash-300">FinDash</p>
          <h1 className="mt-1 text-lg font-semibold text-foreground">Finance Control Panel</h1>
        </div>

        <nav className="flex flex-1 flex-col gap-2 p-4">
          <NavLink to="/dashboard" className={navClass}>
            Dashboard
          </NavLink>
          <NavLink to="/entries" className={navClass}>
            <WalletCards className="mr-1.5 inline-block h-4 w-4" />
            Entries
          </NavLink>
          <RoleGuard allowedRoles={["admin", "manager"]}>
            <NavLink to="/analytics" className={navClass}>
              <PieChart className="mr-1.5 inline-block h-4 w-4" />
              Analytics
            </NavLink>
          </RoleGuard>
          <RoleGuard allowedRoles={["admin", "manager"]}>
            <NavLink to="/users" className={navClass}>
              <Users className="mr-1.5 inline-block h-4 w-4" />
              Users
            </NavLink>
          </RoleGuard>
        </nav>

        <div className="space-y-3 border-t border-border/80 p-4">
          <div>
            <p className="text-sm font-medium text-foreground">{user?.name}</p>
            <p className="text-xs text-muted-foreground">{user?.email}</p>
            <div className="mt-2">
              <Badge tone={user?.role}>{user?.role}</Badge>
            </div>
          </div>
          <Button variant="secondary" size="sm" onClick={logout}>
            <LogOut className="h-4 w-4" />
            Logout
          </Button>
        </div>
      </aside>
    </>
  );
}
