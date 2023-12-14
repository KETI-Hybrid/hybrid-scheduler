#! /bin/bash

for idx in {1..30}
do
	filename='test'
	filename=$filename${idx}
	filename=$filename'.yaml'
	echo ${filename}
	k delete -f ${filename}
done	
