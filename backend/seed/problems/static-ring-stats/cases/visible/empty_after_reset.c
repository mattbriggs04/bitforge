stats_reset();
case_passed = stats_count() == 0 &&
              stats_min() == 0 &&
              stats_max() == 0 &&
              stats_avg() == 0;
