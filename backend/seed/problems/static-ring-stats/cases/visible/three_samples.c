stats_reset();
stats_record(10);
stats_record(30);
stats_record(20);
case_passed = stats_count() == 3 &&
              stats_min() == 10 &&
              stats_max() == 30 &&
              stats_avg() == 20;
