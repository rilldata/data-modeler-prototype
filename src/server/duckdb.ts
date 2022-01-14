/**
 * A single-process duckdb engine.
 */


// @ts-nocheck
import fs from "fs";
import duckdb from 'duckdb';
import { default as glob } from 'glob';

import { guidGenerator } from "../util/guid.js";
// import { Piscina } from "piscina";

interface DB {
	all: Function;
	exec: Function;
	run: Function
}

export function connect() : DB {
	return new duckdb.Database(':memory:');
}

const db:DB = connect();

let onCallback;
let offCallback;

/** utilize these for setting the "running" and "not running" state in the frontend */
export function registerDBRunCallbacks(onCall:Function, offCall:Function) {
	onCallback = onCall;
	offCallback = offCall;
}

function dbAll(db:DB, query:string) {
	if (onCallback) {
		onCallback();
	}
	return new Promise((resolve, reject) => {
		try {
			db.all(query, (err, res) => {
				if (err !== null) {
					reject(err);
				} else {
					if (offCallback) offCallback();
					resolve(res);
				}
			});
		} catch (err) {
			reject(err);
		}
	});
};

export function dbRun(query:string) { 
	return new Promise((resolve, reject) => {
		db.run(query, (err) => {
				if (err !== null) reject(false);
				resolve(true);
			}
		)
	})
}

export async function validQuery(db:DB, query:string): Promise<{value: boolean, message?: string}> {
	return new Promise((resolve) => {
		db.run(query, (err) => {
			if (err !== null) {
				resolve({
					value: false,
					message: err.message
				});
			} else {
				resolve({ value: true});
			}
		});
	});
}

export function hasCreateStatement(query:string) {
	return query.toLowerCase().startsWith('create')
		? `Query has a CREATE statement. 
	Let us handle that for you!
	Just use SELECT and we'll do the rest.
	`
		: false;
}

export function containsMultipleQueries(query:string) {
	return query.split().filter((character) => character == ';').length > 1
		? 'Keep it to a single query please!'
		: false;
}

export function validateQuery(query:string, ...validators:Function[]) {
	return validators.map((validator) => validator(query)).filter((validation) => validation);
}

function wrapQueryAsTemporaryView(query:string) {
	return `CREATE OR REPLACE TEMPORARY VIEW tmp AS (
	${query.replace(';', '')}
);`;
}

export async function checkQuery(query:string) : Promise<void> {
	const output = {};
	const isValid = await validQuery(db, query);
	if (!(isValid.value)) {
		throw Error(isValid.message);
	}
	const validation = validateQuery(query, hasCreateStatement, containsMultipleQueries);
	if (validation.length) {
		throw Error(validation[0])
	}
}

export async function wrapQueryAsView(query:string) {
	return new Promise((resolve, reject) => {
		db.run(wrapQueryAsTemporaryView(query), (err) => {
			if (err !== null) reject(err);
			resolve(true);
		})
	})
}

export async function createPreview(query:string) {
    // FIXME: sort out the type here
	let preview:any;
    try {
		try {
			// get the preview.
			preview = await dbAll(db, 'SELECT * from tmp LIMIT 25;');
		} catch (err) {
			throw Error(err);
		}
	} catch (err) {
		throw Error(err)
	}
    return preview;
}

export async function createSourceProfile(parquetFile:string) {
	return await dbAll(db, `select * from parquet_schema('${parquetFile}');`) as any[];
}

export async function parquetToDBTypes(parquetFile:string) {
	const guid = guidGenerator().replace(/-/g, '_');
    await dbAll(db, `
	CREATE TEMP TABLE tbl_${guid} AS (
        SELECT * from '${parquetFile}' LIMIT 1
    );
	`);
	const tableDef = await dbAll(db, `PRAGMA table_info(tbl_${guid});`)
	await dbAll(db, `DROP TABLE tbl_${guid};`);
    return tableDef;
}

export async function getCardinality(parquetFile:string) {
	const [cardinality] =  await dbAll(db, `select count(*) as count FROM '${parquetFile}';`);
	return cardinality.count;
}

export async function getFirstN(table, n=1) {
	return  dbAll(db, `SELECT * from ${table} LIMIT ${n};`);
}

export function extractParquetFilesFromQuery(query:string) {
	let re = /'[^']*\.parquet'/g;
	const matches = query.match(re);
	if (matches === null) { return null };
	return matches.map(match => match.replace(/'/g, ''));;
}

export async function createSourceProfileFromQuery(query:string) {
	// capture output from parquet query.
	const matches = extractParquetFilesFromQuery(query);
	const tables = (matches === null) ? [] : await Promise.all(matches.map(async (strippedMatch) => {
		//let strippedMatch = match.replace(/'/g, '');
		let match = `'${strippedMatch}'`;
		const info = await createSourceProfile(strippedMatch);
		const head = await getFirstN(match);
		const cardinality = await getCardinality(strippedMatch);
		const sizeInBytes = await getDestinationSize(strippedMatch);
		return {
			profile: info.filter(i => i.name !== 'duckdb_schema'),
			head, 
			cardinality,
			table: strippedMatch,
			sizeInBytes,
			path: strippedMatch,
			name: strippedMatch.split('/').slice(-1)[0]
		}
	}))
	return tables;
}

export async function getDestinationSize(path:string) {
	if (fs.existsSync(path)) {
		const size = await dbAll(db, `SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
		return size.reduce((acc:number, v:object) => acc + v.total_compressed_size, 0)
	}
	return undefined;
}

export async function calculateDestinationCardinality(query:string) {
	const [outputSize] = await dbAll(db, 'SELECT count(*) AS cardinality from tmp;') as any[];
	return outputSize.cardinality;
}

export async function createDestinationProfile(query:string) {
	const info = await dbAll(db, `PRAGMA table_info(tmp);`);
	return info;
}

export async function exportToParquet(query:string, output:string) {
	// generate export just in case.
	if (!fs.existsSync('./export')) {
		fs.mkdirSync('./export');
	}
	const exportQuery = `COPY (${query.replace(';', '')}) TO '${output}' (FORMAT 'parquet')`;
	return dbRun(exportQuery);
}

export async function getParquetFilesInRoot() {
	return new Promise((resolve, reject) => {
		glob.glob('./**/*.parquet', {ignore: ['./node_modules/', './.svelte-kit/', './build/', './src/', './tsc-tmp']},
			(err, output) => {
				if (err!==null) reject(err);
				resolve(output);
			}
		)
	});
}
/**
 * getSummary
 * number: five number summary + mean
 * date: max, min, total time between the two
 * categorical: cardinality
 */

//  export function toDistributionSummary(column:string) {
// 	return [
// 		`min(${column}) as min_${column}`,
// 		`approx_quantile(${column}, 0.25) as q25_${column}`,
// 		`approx_quantile(${column}, 0.5)  as q50_${column}`,
// 		`approx_quantile(${column}, 0.75) as q75_${column}`,
// 		`max(${column}) as max_${column}`,
// 		`avg(${column}) as mean_${column}`,
// 		`stddev_pop(${column}) as sd_${column}`,
// 	]
// }

// // FIXME: deprecate and remove all code paths
// export async function getDistributionSummary(parquetFilePath:string, column:string) {
// 	const [point] = await dbAll(db, `
// SELECT 
// 	min(${column}) as min, 
// 	approx_quantile(${column}, 0.25) as q25, 
// 	approx_quantile(${column}, 0.5)  as q50,
// 	approx_quantile(${column}, 0.75) as q75,
// 	max(${column}) as max,
// 	avg(${column}) as mean,
// 	stddev_pop(${column}) as sd
// 	FROM '${parquetFilePath}';`);
// 	return point;
// }










export function toDistributionSummary(column) {
	return [
		`min(${column}) as ${column}_min`,
		`reservoir_quantile(${column}, 0.25) as ${column}_q25`,
		`reservoir_quantile(${column}, 0.5)  as ${column}_q50`,
		`reservoir_quantile(${column}, 0.75) as ${column}_q75`,
		`max(${column}) as ${column}_max`,
		`avg(${column}) as ${column}_mean`,
		`stddev_pop(${column}) as ${column}_sd`,
	]
}

// const piscina = new Piscina({
// 	filename: new URL('./duckdb-worker.js', import.meta.url).href
// })


function topK(tablePath, column) {
	return `SELECT ${column} as value, count(*) AS count from ${tablePath}
GROUP BY ${column}
ORDER BY count desc
LIMIT 50;`
}


export async function getTopKAndCardinality(tablePath, column, dbEngine = db) {
	const topKValues = await dbAll(dbEngine, topK(tablePath, column));
	const [cardinality] = await dbAll(dbEngine = db, `SELECT approx_count_distinct(${column}) as count from ${tablePath};`);
	return {
		column,
		topK: topKValues,
		cardinality: cardinality.count
	}
}

export async function getNullCount(tablePath:string, field:string, dbEngine = db) {
	const [nullity] = await dbAll(dbEngine, `
		SELECT COUNT(*) as count FROM ${tablePath} WHERE ${field} IS NULL;
	`);
	return nullity.count;
}

export async function getNullCounts(tablePath:string, fields:any, dbEngine = db) {
	const [nullities] = await dbAll(dbEngine, `
		SELECT
		${fields.map(field => {
			return `COUNT(CASE WHEN ${field.name} IS NULL THEN 1 ELSE NULL END) as ${field.name}`
		}).join(',\n')}
		FROM ${tablePath};
	`);
	return nullities;
}

export async function numericHistogram(tablePath:string, field:string, fieldType:string, dbEngine = db) {

	// if the field type is an integer and the total number of values is low, can't we just use
	// first check a sample to see how many buckets there are for this value.
	const buckets = await dbAll(dbEngine, `SELECT count(*) as count, ${field} FROM ${tablePath} WHERE ${field} IS NOT NULL GROUP BY ${field} USING SAMPLE reservoir(1000 ROWS);`)
	const bucketSize = Math.min(40, buckets.length);
	return dbAll(dbEngine, `
	WITH dataset AS (
		SELECT ${fieldType === 'TIMESTAMP' ? `epoch(${field})` : `${field}::DOUBLE`} as ${field} FROM ${tablePath}
	) , S AS (
		SELECT 
			min(${field}) as minVal,
			max(${field}) as maxVal,
			(max(${field}) - min(${field})) as range
			FROM dataset
	), values AS (
		SELECT ${field} as value from dataset
		WHERE ${field} IS NOT NULL
	), buckets AS (
		SELECT
			range as bucket,
			(range - 1) * (select range FROM S) / ${bucketSize} + (select minVal from S) as low,
			(range) * (select range FROM S) / ${bucketSize} + (select minVal from S) as high
		FROM range(0, ${bucketSize}, 1)
	)
	, histogram_stage AS (
		SELECT
			bucket,
			low,
			high,
			count(values.value) as count
		FROM buckets
		LEFT JOIN values ON values.value BETWEEN low and high
		GROUP BY bucket, low, high
		ORDER BY BUCKET
	)
	SELECT 
		bucket,
		low,
		high,
		CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
		FROM histogram_stage;
	
	`)
}