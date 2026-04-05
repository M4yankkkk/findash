import { useCallback, useEffect, useState } from "react";
import api from "@/api/axios";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useAuth } from "@/hooks/useAuth";

function formatCurrency(value) {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 2 }).format(value || 0);
}

export default function Entries() {
  const { user } = useAuth();
  const canCreateEntry = user?.role === "admin" || user?.role === "manager";

  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [category, setCategory] = useState("");
  const [entries, setEntries] = useState([]);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [createError, setCreateError] = useState("");
  const [createSuccess, setCreateSuccess] = useState("");
  const [form, setForm] = useState({
    title: "",
    amount: "",
    type: "expense",
    category: "",
    description: "",
    date: new Date().toISOString().split("T")[0],
  });

  const fetchEntries = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.get("/entries", {
        params: {
          page,
          page_size: pageSize,
          category: category || undefined,
        },
      });

      const payload = response.data?.data;
      setEntries(payload?.items || []);
      setTotal(payload?.total || 0);
      setHasMore(Boolean(payload?.has_more));
    } catch {
      setEntries([]);
      setTotal(0);
      setHasMore(false);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, category]);

  useEffect(() => {
    fetchEntries();
  }, [fetchEntries]);

  const handleCreateEntry = async (event) => {
    event.preventDefault();
    setCreateError("");
    setCreateSuccess("");
    setSubmitting(true);

    try {
      await api.post("/entries", {
        title: form.title,
        amount: Number(form.amount),
        type: form.type,
        category: form.category,
        description: form.description,
        date: `${form.date}T00:00:00Z`,
      });

      setCreateSuccess("Entry added successfully.");
      setForm((prev) => ({
        ...prev,
        title: "",
        amount: "",
        category: "",
        description: "",
      }));

      if (page !== 1) {
        setPage(1);
      } else {
        fetchEntries();
      }
    } catch (error) {
      setCreateError(error.response?.data?.error || "Failed to add entry.");
    } finally {
      setSubmitting(false);
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Entries</h2>
          <p className="text-sm text-muted-foreground">Search entries by category with server-side pagination.</p>
        </div>

        <div className="w-full sm:w-80">
          <label className="mb-2 block text-sm text-muted-foreground">Search Category</label>
          <Input
            value={category}
            onChange={(event) => {
              setPage(1);
              setCategory(event.target.value);
            }}
            placeholder="e.g. salary, food, rent"
          />
        </div>
      </div>

      {canCreateEntry ? (
        <Card className="border-ash-700/50 bg-ash-900/70">
          <CardHeader>
            <CardTitle>Add New Entry</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreateEntry} className="grid gap-4 lg:grid-cols-2">
              <div>
                <label className="mb-2 block text-sm text-muted-foreground">Title</label>
                <Input
                  required
                  value={form.title}
                  onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
                  placeholder="April rent"
                />
              </div>

              <div>
                <label className="mb-2 block text-sm text-muted-foreground">Category</label>
                <Input
                  required
                  value={form.category}
                  onChange={(event) => setForm((prev) => ({ ...prev, category: event.target.value }))}
                  placeholder="housing"
                />
              </div>

              <div>
                <label className="mb-2 block text-sm text-muted-foreground">Amount</label>
                <Input
                  required
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={form.amount}
                  onChange={(event) => setForm((prev) => ({ ...prev, amount: event.target.value }))}
                  placeholder="1000"
                />
              </div>

              <div>
                <label className="mb-2 block text-sm text-muted-foreground">Type</label>
                <select
                  className="h-10 w-full rounded-md border border-input bg-background/40 px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  value={form.type}
                  onChange={(event) => setForm((prev) => ({ ...prev, type: event.target.value }))}
                >
                  <option value="income">Income</option>
                  <option value="expense">Expense</option>
                </select>
              </div>

              <div>
                <label className="mb-2 block text-sm text-muted-foreground">Date</label>
                <Input
                  required
                  type="date"
                  value={form.date}
                  onChange={(event) => setForm((prev) => ({ ...prev, date: event.target.value }))}
                />
              </div>

              <div className="lg:col-span-2">
                <label className="mb-2 block text-sm text-muted-foreground">Description</label>
                <textarea
                  className="min-h-20 w-full rounded-md border border-input bg-background/40 px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  value={form.description}
                  onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
                  placeholder="Optional description"
                />
              </div>

              <div className="flex items-center gap-3 lg:col-span-2">
                <Button type="submit" disabled={submitting}>
                  {submitting ? "Adding..." : "Add Entry"}
                </Button>
                {createSuccess ? <p className="text-sm text-emerald-300">{createSuccess}</p> : null}
                {createError ? <p className="text-sm text-rose-300">{createError}</p> : null}
              </div>
            </form>
          </CardContent>
        </Card>
      ) : (
        <Card className="border-ash-700/50 bg-ash-900/70">
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Viewers can browse entries but cannot add new ones.</p>
          </CardContent>
        </Card>
      )}

      <Card className="border-ash-700/50 bg-ash-900/70">
        <CardHeader>
          <CardTitle>Financial Entries</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Date</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {entries.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell>{entry.title}</TableCell>
                  <TableCell>{entry.category}</TableCell>
                  <TableCell>
                    <Badge tone={entry.type}>{entry.type}</Badge>
                  </TableCell>
                  <TableCell>{new Date(entry.date).toLocaleDateString()}</TableCell>
                  <TableCell className="text-right">{formatCurrency(entry.amount)}</TableCell>
                </TableRow>
              ))}

              {!entries.length && !loading ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    No entries found.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm text-muted-foreground">
              Page {page} of {totalPages} • {total} total entries
            </p>
            <div className="flex gap-2">
              <Button variant="secondary" onClick={() => setPage((prev) => Math.max(1, prev - 1))} disabled={page === 1 || loading}>
                Previous
              </Button>
              <Button variant="secondary" onClick={() => setPage((prev) => prev + 1)} disabled={!hasMore || loading}>
                Next
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
