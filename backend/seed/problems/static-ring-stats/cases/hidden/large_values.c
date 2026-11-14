stats_reset();
stats_record(60000);
stats_record(65535);
stats_record(65535);
case_passed = stats_count() == 3 &&
              stats_min() == 60000 &&
              stats_max() == 65535 &&
              stats_avg() == 63690;
