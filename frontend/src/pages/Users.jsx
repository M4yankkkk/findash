import { useEffect, useState } from "react";
import api from "@/api/axios";
import { useAuth } from "@/hooks/useAuth";
import RoleGuard from "@/components/RoleGuard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const roles = ["viewer", "manager", "admin"];

export default function Users() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [users, setUsers] = useState([]);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [updatingId, setUpdatingId] = useState("");

  const [viewers, setViewers] = useState([]);
  const [selectedViewerID, setSelectedViewerID] = useState("");
  const [assignedEntryIDs, setAssignedEntryIDs] = useState([]);
  const [availableEntries, setAvailableEntries] = useState([]);
  const [savingVisibility, setSavingVisibility] = useState(false);
  const [visibilityMessage, setVisibilityMessage] = useState("");
  const [visibilityError, setVisibilityError] = useState("");

  useEffect(() => {
    if (!isAdmin) {
      return;
    }

    const fetchUsers = async () => {
      try {
        const response = await api.get("/users", { params: { page, page_size: pageSize } });
        const payload = response.data?.data;
        setUsers(payload?.items || []);
        setTotal(payload?.total || 0);
        setHasMore(Boolean(payload?.has_more));
      } catch {
        setUsers([]);
        setTotal(0);
        setHasMore(false);
      }
    };

    fetchUsers();
  }, [page, pageSize, isAdmin]);

  useEffect(() => {
    const fetchViewers = async () => {
      try {
        const response = await api.get("/viewer-visibility/viewers", { params: { page: 1, page_size: 100 } });
        const list = response.data?.data?.items || [];
        setViewers(list);
        if (list.length > 0) {
          setSelectedViewerID((prev) => prev || list[0].id);
        }
      } catch {
        setViewers([]);
      }
    };

    fetchViewers();
  }, []);

  useEffect(() => {
    const fetchAssignableEntries = async () => {
      try {
        const response = await api.get("/entries", { params: { page: 1, page_size: 100 } });
        setAvailableEntries(response.data?.data?.items || []);
      } catch {
        setAvailableEntries([]);
      }
    };

    fetchAssignableEntries();
  }, []);

  useEffect(() => {
    if (!selectedViewerID) {
      setAssignedEntryIDs([]);
      return;
    }

    const fetchViewerVisibility = async () => {
      try {
        const response = await api.get(`/viewer-visibility/${selectedViewerID}/entries`);
        setAssignedEntryIDs(response.data?.data?.entry_ids || []);
      } catch {
        setAssignedEntryIDs([]);
      }
    };

    fetchViewerVisibility();
  }, [selectedViewerID]);

  const handleRoleChange = async (userId, role) => {
    setUpdatingId(userId);
    try {
      await api.patch(`/users/${userId}/role`, { role });
      setUsers((prev) => prev.map((u) => (u.id === userId ? { ...u, role } : u)));
    } finally {
      setUpdatingId("");
    }
  };

  const toggleAssignedEntry = (entryID) => {
    setAssignedEntryIDs((prev) =>
      prev.includes(entryID) ? prev.filter((id) => id !== entryID) : [...prev, entryID]
    );
  };

  const handleSaveVisibility = async () => {
    if (!selectedViewerID) {
      return;
    }

    setSavingVisibility(true);
    setVisibilityMessage("");
    setVisibilityError("");
    try {
      await api.put(`/viewer-visibility/${selectedViewerID}/entries`, { entry_ids: assignedEntryIDs });
      setVisibilityMessage("Viewer visibility updated.");
    } catch (error) {
      setVisibilityError(error.response?.data?.error || "Failed to update visibility.");
    } finally {
      setSavingVisibility(false);
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const selectedViewer = viewers.find((viewer) => viewer.id === selectedViewerID);

  return (
    <RoleGuard
      allowedRoles={["admin", "manager"]}
      fallback={<p className="text-sm text-muted-foreground">Users page is available for manager and admin roles only.</p>}
    >
      <section className="space-y-6">
        <div>
          <h2 className="text-2xl font-semibold">Users</h2>
          <p className="text-sm text-muted-foreground">Assign what entries each viewer can see.</p>
        </div>

        <Card className="border-ash-700/50 bg-ash-900/70">
          <CardHeader>
            <CardTitle>Viewer Entry Visibility</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="max-w-md">
              <label className="mb-2 block text-sm text-muted-foreground">Select Viewer</label>
              <select
                className="h-10 w-full rounded-md border border-input bg-background/40 px-3 py-2 text-sm text-foreground"
                value={selectedViewerID}
                onChange={(event) => setSelectedViewerID(event.target.value)}
              >
                {viewers.map((viewer) => (
                  <option key={viewer.id} value={viewer.id}>
                    {viewer.name} ({viewer.email})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <p className="mb-2 text-sm text-muted-foreground">
                Assign visible entries for {selectedViewer ? `${selectedViewer.name}` : "selected viewer"}
              </p>
              <div className="max-h-72 space-y-2 overflow-y-auto rounded-md border border-border p-3">
                {availableEntries.map((entry) => {
                  const checked = assignedEntryIDs.includes(entry.id);
                  return (
                    <label key={entry.id} className="flex items-start gap-3 rounded-md px-2 py-2 hover:bg-muted/20">
                      <input
                        type="checkbox"
                        className="mt-1"
                        checked={checked}
                        onChange={() => toggleAssignedEntry(entry.id)}
                      />
                      <div>
                        <p className="text-sm font-medium">{entry.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {entry.category} • {entry.type} • {new Date(entry.date).toLocaleDateString()}
                        </p>
                      </div>
                    </label>
                  );
                })}
                {!availableEntries.length ? <p className="text-sm text-muted-foreground">No entries available to assign.</p> : null}
              </div>
            </div>

            <div className="flex items-center gap-3">
              <Button onClick={handleSaveVisibility} disabled={savingVisibility || !selectedViewerID}>
                {savingVisibility ? "Saving..." : "Save Visibility"}
              </Button>
              {visibilityMessage ? <p className="text-sm text-emerald-300">{visibilityMessage}</p> : null}
              {visibilityError ? <p className="text-sm text-rose-300">{visibilityError}</p> : null}
            </div>
          </CardContent>
        </Card>

        {isAdmin ? (
          <Card className="border-ash-700/50 bg-ash-900/70">
            <CardHeader>
              <CardTitle>All Users</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead className="text-right">Update Role</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((listedUser) => (
                    <TableRow key={listedUser.id}>
                      <TableCell>{listedUser.name}</TableCell>
                      <TableCell>{listedUser.email}</TableCell>
                      <TableCell>
                        <Badge tone={listedUser.role}>{listedUser.role}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="inline-flex items-center gap-2">
                          <select
                            className="h-9 rounded-md border border-input bg-background/40 px-2 text-sm"
                            value={listedUser.role}
                            onChange={(e) => handleRoleChange(listedUser.id, e.target.value)}
                            disabled={updatingId === listedUser.id}
                          >
                            {roles.map((role) => (
                              <option key={role} value={role}>
                                {role}
                              </option>
                            ))}
                          </select>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {!users.length ? (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center text-muted-foreground">
                        No users found.
                      </TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>

              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">
                  Page {page} of {totalPages} • {total} users
                </p>
                <div className="flex gap-2">
                  <Button variant="secondary" onClick={() => setPage((prev) => Math.max(1, prev - 1))} disabled={page === 1}>
                    Previous
                  </Button>
                  <Button variant="secondary" onClick={() => setPage((prev) => prev + 1)} disabled={!hasMore}>
                    Next
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ) : null}
      </section>
    </RoleGuard>
  );
}
