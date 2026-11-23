ALTER TABLE public.users           ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.scrapes         ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.scrape_entries  ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.scrape_failures ENABLE ROW LEVEL SECURITY;

GRANT USAGE ON SCHEMA public TO anon, authenticated;
GRANT SELECT ON
    public.users,
    public.scrapes,
    public.scrape_entries,
    public.scrape_failures
TO anon, authenticated;

CREATE POLICY "public read" ON public.users           FOR SELECT TO anon, authenticated USING (true);
CREATE POLICY "public read" ON public.scrapes         FOR SELECT TO anon, authenticated USING (true);
CREATE POLICY "public read" ON public.scrape_entries  FOR SELECT TO anon, authenticated USING (true);
CREATE POLICY "public read" ON public.scrape_failures FOR SELECT TO anon, authenticated USING (true);
