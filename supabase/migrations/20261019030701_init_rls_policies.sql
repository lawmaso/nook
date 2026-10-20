ALTER TABLE users
  ENABLE ROW LEVEL SECURITY;
ALTER TABLE scrapes
  ENABLE ROW LEVEL SECURITY;
ALTER TABLE scrape_entries
  ENABLE ROW LEVEL SECURITY;

CREATE policy "public read" ON users FOR SELECT TO anon USING (TRUE);
CREATE policy "public read" ON scrapes FOR SELECT TO anon USING (TRUE);
CREATE policy "public read" ON scrape_entries FOR SELECT TO anon USING (TRUE);
