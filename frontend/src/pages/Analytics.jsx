import { useEffect, useState } from "react";
import { Bar, BarChart, CartesianGrid, Cell, Legend, Line, LineChart, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import api from "@/api/axios";
import RoleGuard from "@/components/RoleGuard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const chartColors = ["#80956a", "#667755", "#99aa88", "#4d5a3f", "#b3c0a5", "#ccd5c3"];

export default function Analytics() {
  const [summary, setSummary] = useState(null);
  const [categoryData, setCategoryData] = useState([]);
  const [trendData, setTrendData] = useState([]);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        const [summaryRes, categoryRes, trendRes] = await Promise.all([
          api.get("/analytics/summary"),
          api.get("/analytics/by-category"),
          api.get("/analytics/trend", { params: { months: 6 } }),
        ]);

        setSummary(summaryRes.data?.data || null);
        setCategoryData(categoryRes.data?.data || []);
        setTrendData(trendRes.data?.data || []);
      } catch {
        setSummary(null);
        setCategoryData([]);
        setTrendData([]);
      }
    };

    fetchAnalytics();
  }, []);

  return (
    <RoleGuard
      allowedRoles={["admin", "manager"]}
      fallback={<p className="text-sm text-muted-foreground">Analytics is available for manager and admin roles only.</p>}
    >
      <section className="space-y-6">
        <div>
          <h2 className="text-2xl font-semibold">Analytics</h2>
          <p className="text-sm text-muted-foreground">Summary, category split, and monthly trend.</p>
        </div>

        <Card className="border-ash-700/50 bg-ash-900/70">
          <CardHeader>
            <CardTitle>Summary</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-3">
            <div className="rounded-md border border-border bg-background/30 p-4">
              <p className="text-xs uppercase tracking-wider text-muted-foreground">Income</p>
              <p className="mt-2 text-lg font-semibold">{summary?.total_income ?? 0}</p>
            </div>
            <div className="rounded-md border border-border bg-background/30 p-4">
              <p className="text-xs uppercase tracking-wider text-muted-foreground">Expenses</p>
              <p className="mt-2 text-lg font-semibold">{summary?.total_expenses ?? 0}</p>
            </div>
            <div className="rounded-md border border-border bg-background/30 p-4">
              <p className="text-xs uppercase tracking-wider text-muted-foreground">Net Balance</p>
              <p className="mt-2 text-lg font-semibold">{summary?.net_balance ?? 0}</p>
            </div>
          </CardContent>
        </Card>

        <div className="grid gap-6 xl:grid-cols-2">
          <Card className="border-ash-700/50 bg-ash-900/70">
            <CardHeader>
              <CardTitle>By Category</CardTitle>
            </CardHeader>
            <CardContent className="h-80">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={categoryData} dataKey="total" nameKey="category" outerRadius={110}>
                    {categoryData.map((_, index) => (
                      <Cell key={`slice-${index}`} fill={chartColors[index % chartColors.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          <Card className="border-ash-700/50 bg-ash-900/70">
            <CardHeader>
              <CardTitle>Monthly Trend</CardTitle>
            </CardHeader>
            <CardContent className="h-80">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={trendData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333c2a" />
                  <XAxis dataKey="month" stroke="#99aa88" />
                  <YAxis stroke="#99aa88" />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="total_income" stroke="#80956a" strokeWidth={2.5} />
                  <Line type="monotone" dataKey="total_expenses" stroke="#b3c0a5" strokeWidth={2.5} />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>

        <Card className="border-ash-700/50 bg-ash-900/70">
          <CardHeader>
            <CardTitle>Income vs Expense</CardTitle>
          </CardHeader>
          <CardContent className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={trendData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#333c2a" />
                <XAxis dataKey="month" stroke="#99aa88" />
                <YAxis stroke="#99aa88" />
                <Tooltip />
                <Legend />
                <Bar dataKey="total_income" fill="#80956a" radius={[6, 6, 0, 0]} />
                <Bar dataKey="total_expenses" fill="#4d5a3f" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </section>
    </RoleGuard>
  );
}
