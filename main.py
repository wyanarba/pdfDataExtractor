aa = "#0F172A #334155 #94A3B8 #F1F5F9 #F8FAFC #2563EB #3B82F6 #60A5FA #8B5CF6 #A78BFA #E11D48 #F43F5E #EA580C #F97316 #F59E0B #059669 #10B981 #34D399 #0D9488 #14B8A6"

for el in aa.split():
    print("ColorNN1 = color.RGBA{" f"0x{el[1:3]}, 0x{el[3:5]}, 0x{el[5:]}" "} //" f"{el}")