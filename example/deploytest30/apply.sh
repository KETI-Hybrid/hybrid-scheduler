#! /bin/bash

for idx in {1..1}
do
	filename='test'
	filename=$filename${idx}
	filename=$filename'.yaml'
	echo ${filename}
	k apply -f ${filename}
done	
