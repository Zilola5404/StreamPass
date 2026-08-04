-- BL-026: normalize legacy region codes to canonical catalog values.
UPDATE relay_servers
SET region = lower(trim(region))
WHERE region IS NOT NULL;

UPDATE relay_servers SET region = 'nl'
WHERE lower(region) IN ('nl', 'netherlands', 'amsterdam', 'holland')
   OR region ILIKE '%amsterdam%'
   OR region ILIKE '%netherlands%';

UPDATE relay_servers SET region = 'de'
WHERE lower(region) IN ('de', 'germany', 'frankfurt')
   OR region ILIKE '%frankfurt%'
   OR region ILIKE '%germany%';

UPDATE relay_servers SET region = 'pl'
WHERE lower(region) IN ('pl', 'poland', 'warsaw', 'warszawa')
   OR region ILIKE '%warsaw%'
   OR region ILIKE '%poland%';

UPDATE relay_servers SET region = 'fi'
WHERE lower(region) IN ('fi', 'finland', 'helsinki')
   OR region ILIKE '%helsinki%'
   OR region ILIKE '%finland%';
