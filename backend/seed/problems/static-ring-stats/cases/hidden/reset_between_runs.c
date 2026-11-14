stats_reset();
stats_record(100);
stats_record(200);
stats_reset();
stats_record(7);
case_passed = stats_count() == 1 &&
              stats_min() == 7 &&
              stats_max() == 7 &&
              stats_avg() == 7;
