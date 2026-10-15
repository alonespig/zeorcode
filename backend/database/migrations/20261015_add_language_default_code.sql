ALTER TABLE `languages`
  ADD COLUMN `default_code` text NULL AFTER `sort`;

UPDATE `languages`
SET `default_code` = CASE LOWER(`name`)
  WHEN 'c++' THEN CONCAT(
    '#include <bits/stdc++.h>', CHAR(10),
    'using namespace std;', CHAR(10), CHAR(10),
    'int main() {', CHAR(10),
    '    ios::sync_with_stdio(false);', CHAR(10),
    '    cin.tie(nullptr);', CHAR(10), CHAR(10),
    '    return 0;', CHAR(10),
    '}', CHAR(10)
  )
  WHEN 'java' THEN CONCAT(
    'import java.io.*;', CHAR(10),
    'import java.util.*;', CHAR(10), CHAR(10),
    'public class Main {', CHAR(10),
    '    public static void main(String[] args) throws Exception {', CHAR(10), CHAR(10),
    '    }', CHAR(10),
    '}', CHAR(10)
  )
  WHEN 'python' THEN CONCAT(
    'import sys', CHAR(10), CHAR(10),
    'def solve():', CHAR(10),
    '    pass', CHAR(10), CHAR(10),
    'if __name__ == "__main__":', CHAR(10),
    '    solve()', CHAR(10)
  )
  ELSE ''
END
WHERE `default_code` IS NULL OR `default_code` = '';

ALTER TABLE `languages`
  MODIFY COLUMN `default_code` text NOT NULL AFTER `sort`;
