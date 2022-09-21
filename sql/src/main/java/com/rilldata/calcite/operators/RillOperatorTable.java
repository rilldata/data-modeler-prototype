package com.rilldata.calcite.operators;

import com.rilldata.calcite.utility.CalciteUtils;
import org.apache.calcite.sql.SqlFunction;
import org.apache.calcite.sql.SqlFunctionCategory;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlOperator;
import org.apache.calcite.sql.SqlOperatorTable;
import org.apache.calcite.sql.SqlSyntax;
import org.apache.calcite.sql.fun.SqlLibrary;
import org.apache.calcite.sql.fun.SqlLibraryOperatorTableFactory;
import org.apache.calcite.sql.fun.SqlStdOperatorTable;
import org.apache.calcite.sql.type.OperandTypes;
import org.apache.calcite.sql.type.ReturnTypes;
import org.apache.calcite.sql.type.SqlTypeFamily;
import org.apache.calcite.sql.util.ListSqlOperatorTable;
import org.apache.calcite.sql.validate.SqlNameMatcher;
import org.checkerframework.checker.nullness.qual.Nullable;

import java.util.ArrayList;
import java.util.EnumSet;
import java.util.List;

public class RillOperatorTable implements SqlOperatorTable
{
  private static final SqlFunction DATE_TRUNC =
    new SqlFunction(
        "DATE_TRUNC",
        SqlKind.OTHER_FUNCTION,
        ReturnTypes.TIMESTAMP,
        null,
        OperandTypes.family(SqlTypeFamily.CHARACTER, SqlTypeFamily.TIMESTAMP),
        SqlFunctionCategory.TIMEDATE
    );

  private static final SqlFunction GREATEST =
      new SqlFunction(
          "GREATEST",
          SqlKind.GREATEST,
          CalciteUtils.TYPE_INFERENCE,
          null,
          OperandTypes.VARIADIC,
          SqlFunctionCategory.SYSTEM
      );

  private static final SqlFunction LEAST =
      new SqlFunction(
          "LEAST",
          SqlKind.LEAST,
          CalciteUtils.TYPE_INFERENCE,
          null,
          OperandTypes.VARIADIC,
          SqlFunctionCategory.SYSTEM
      );

  public static final SqlFunction LOG =
      new SqlFunction(
          "LOG",
          SqlKind.OTHER_FUNCTION,
          ReturnTypes.DOUBLE_NULLABLE,
          null,
          OperandTypes.NUMERIC,
          SqlFunctionCategory.NUMERIC
      );

  public static final SqlFunction LOG2 =
      new SqlFunction(
          "LOG2",
          SqlKind.OTHER_FUNCTION,
          ReturnTypes.DOUBLE_NULLABLE,
          null,
          OperandTypes.NUMERIC,
          SqlFunctionCategory.NUMERIC
      );

  // This is called BITWISE_XOR in Druid
  public static final SqlFunction XOR =
      new SqlFunction(
          "XOR",
          SqlKind.OTHER_FUNCTION,
          ReturnTypes.INTEGER,
          null,
          OperandTypes.family(SqlTypeFamily.INTEGER, SqlTypeFamily.INTEGER),
          SqlFunctionCategory.NUMERIC
      );
  private final List<SqlOperatorTable> operatorTables;

  public RillOperatorTable()
  {
    operatorTables = new ArrayList<>();
    // Add standard sql operators
    SqlStdOperatorTable sqlStdOperatorTable = SqlStdOperatorTable.instance();
    operatorTables.add(sqlStdOperatorTable);
    // Add postgres operators
    SqlOperatorTable postgresOperators = SqlLibraryOperatorTableFactory.INSTANCE.getOperatorTable(
        EnumSet.of(SqlLibrary.POSTGRESQL));
    operatorTables.add(postgresOperators);

    // add custom operators
    final ListSqlOperatorTable customOperatorTable = new ListSqlOperatorTable();
    customOperatorTable.add(DATE_TRUNC);
    customOperatorTable.add(GREATEST);
    customOperatorTable.add(LEAST);
    customOperatorTable.add(LOG);
    customOperatorTable.add(LOG2);
    customOperatorTable.add(XOR);
    operatorTables.add(customOperatorTable);
  }

  @Override public void lookupOperatorOverloads(SqlIdentifier opName, @Nullable SqlFunctionCategory category,
      SqlSyntax syntax, List<SqlOperator> operatorList, SqlNameMatcher nameMatcher
  )
  {
    for (SqlOperatorTable table : operatorTables) {
      table.lookupOperatorOverloads(opName, category, syntax, operatorList, nameMatcher);
    }
  }

  @Override public List<SqlOperator> getOperatorList()
  {
    List<SqlOperator> list = new ArrayList<>();
    for (SqlOperatorTable table : operatorTables) {
      list.addAll(table.getOperatorList());
    }
    return list;
  }
}
