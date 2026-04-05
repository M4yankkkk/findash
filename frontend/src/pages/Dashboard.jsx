import { useEffect, useState } from "react";
import api from "@/api/axios";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

function formatCurrency(value) {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR", maximumFractionDigits: 2 }).format(value || 0);
}

export default function Dashboard() {
  const [summary, setSummary] = useState(null);
  const [recentEntries, setRecentEntries] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [summaryRes, entriesRes] = await Promise.all([
          api.get("/dashboard/summary"),
          api.get("/entries", { params: { page: 1, page_size: 5 } }),
        ]);

        setSummary(summaryRes.data?.data || null);
        setRecentEntries(entriesRes.data?.data?.items || []);
      } catch {
        setSummary(null);
        setRecentEntries([]);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const statCards = [
    { label: "Total Income", value: formatCurrency(summary?.total_income), tone: "income" },
    { label: "Total Expenses", value: formatCurrency(summary?.total_expenses), tone: "expense" },
    { label: "Net Balance", value: formatCurrency(summary?.net_balance), tone: "viewer" },
    { label: "Entry Count", value: summary?.entry_count ?? 0, tone: "manager" },
  ];

  return (
    <section className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Dashboard</h2>
        <p className="text-sm text-muted-foreground">Quick snapshot of your financial activity.</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {statCards.map((card) => (
          <Card key={card.label} className="border-ash-700/50 bg-ash-900/70">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">{card.label}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-xl font-semibold">{loading ? "..." : card.value}</p>
                <Badge tone={card.tone}>{card.tone}</Badge>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="border-ash-700/50 bg-ash-900/70">
        <CardHeader>
          <CardTitle>Recent Entries</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Type</TableHead>
                <TableHead className="text-right">Amount</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {recentEntries.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell>{entry.title}</TableCell>
                  <TableCell>{entry.category}</TableCell>
                  <TableCell>
                    <Badge tone={entry.type}>{entry.type}</Badge>
                  </TableCell>
                  <TableCell className="text-right">{formatCurrency(entry.amount)}</TableCell>
                </TableRow>
              ))}
              {!recentEntries.length ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    No entries yet.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </section>
  );
}
