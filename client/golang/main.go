package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/apache/spark-connect-go/v35/spark/sql"
)

var (
	remote = flag.String("remote", "sc://localhost:15002",
		"the remote address of Spark Connect server to connect to")

	filedir = flag.String("filedir", "/tmp",
		"the directory to save the files")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	spark, err := sql.NewSessionBuilder().Remote(*remote).Build(ctx)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}
	defer spark.Stop()

	df, err := spark.Sql(ctx, "select 'apple' as word, 123 as count union all select 'orange' as word, 456 as count")
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	log.Printf("DataFrame from sql: select 'apple' as word, 123 as count union all select 'orange' as word, 456 as count")
	err = df.Show(ctx, 100, false)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	schema, err := df.Schema(ctx)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	for _, f := range schema.Fields {
		log.Printf("Field in dataframe schema: %s - %s", f.Name, f.DataType.TypeName())
	}

	rows, err := df.Collect(ctx)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	schema, err = rows[0].Schema()
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	for _, f := range schema.Fields {
		log.Printf("Field in row: %s - %s", f.Name, f.DataType.TypeName())
	}

	for _, row := range rows {
		log.Printf("Row: %v", row)
	}

	err = df.Writer().Mode("overwrite").
		Format("parquet").
		Save(ctx, fmt.Sprintf("file://%s/spark-connect-write-example-output.parquet", *filedir))
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	df, err = spark.Read().Format("parquet").
		Load(fmt.Sprintf("file://%s/spark-connect-write-example-output.parquet", *filedir))
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	log.Printf("DataFrame from reading parquet")
	err = df.Show(ctx, 100, false)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	err = df.CreateTempView(ctx, "view1", true, false)
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	df, err = spark.Sql(ctx, "select count, word from view1 order by count")
	if err != nil {
		log.Fatalf("Failed: %s", err)
	}

	log.Printf("DataFrame from sql: select count, word from view1 order by count")
	df.Show(ctx, 100, false)
}
